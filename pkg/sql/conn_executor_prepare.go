// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package sql

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/errors"
	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/pgcode"
	"github.com/lib/pq/oid"
)

func (ex *connExecutor) execPrepare(
	ctx context.Context, parseCmd PrepareStmt,
) (fsm.Event, fsm.EventPayload) {

	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		return eventNonRetriableErr{IsCommit: fsm.False}, eventNonRetriableErrPayload{err: err}
	}

	// The anonymous statement can be overwritten.
	if parseCmd.Name != "" {
		if _, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[parseCmd.Name]; ok {
			err := pgerror.Newf(
				pgcode.DuplicatePreparedStatement,
				"prepared statement %q already exists", parseCmd.Name,
			)
			return retErr(err)
		}
	} else {
		// Deallocate the unnamed statement, if it exists.
		ex.deletePreparedStmt(ctx, "")
	}

	ps, err := ex.addPreparedStmt(
		ctx, parseCmd.Name, Statement{Statement: parseCmd.Statement}, parseCmd.TypeHints,
	)
	if err != nil {
		return retErr(err)
	}

	// Convert the inferred SQL types back to an array of pgwire Oids.
	if len(ps.TypeHints) > pgwirebase.MaxPreparedStatementArgs {
		return retErr(
			pgwirebase.NewProtocolViolationErrorf(
				"more than %d arguments to prepared statement: %d",
				pgwirebase.MaxPreparedStatementArgs, len(ps.TypeHints)))
	}
	inferredTypes := make([]oid.Oid, len(ps.Types))
	copy(inferredTypes, parseCmd.RawTypeHints)

	for i := range ps.Types {
		// OID to Datum is not a 1-1 mapping (for example, int4 and int8
		// both map to TypeInt), so we need to maintain the types sent by
		// the client.
		if inferredTypes[i] == 0 {
			t, _ := ps.ValueType(tree.PlaceholderIdx(i))
			inferredTypes[i] = t.Oid()
		}
	}
	// Remember the inferred placeholder types so they can be reported on
	// Describe.
	ps.InferredTypes = inferredTypes
	return nil, nil
}

// addPreparedStmt creates a new PreparedStatement with the provided name using
// the given query. The new prepared statement is added to the connExecutor and
// also returned. It is illegal to call this when a statement with that name
// already exists (even for anonymous prepared statements).
//
// placeholderHints are used to assist in inferring placeholder types.
func (ex *connExecutor) addPreparedStmt(
	ctx context.Context, name string, stmt Statement, placeholderHints tree.PlaceholderTypes,
) (*PreparedStatement, error) {
	if _, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]; ok {
		panic(fmt.Sprintf("prepared statement already exists: %q", name))
	}

	// Prepare the query. This completes the typing of placeholders.
	prepared, err := ex.prepare(ctx, stmt, placeholderHints)
	if err != nil {
		return nil, err
	}

	if err := prepared.memAcc.Grow(ctx, int64(len(name))); err != nil {
		return nil, err
	}
	ex.extraTxnState.prepStmtsNamespace.prepStmts[name] = prepared
	return prepared, nil
}

// prepare prepares the given statement.
//
// placeholderHints may contain partial type information for placeholders.
// prepare will populate the missing types. It can be nil.
//
// The PreparedStatement is returned (or nil if there are no results). The
// returned PreparedStatement needs to be close()d once its no longer in use.
func (ex *connExecutor) prepare(
	ctx context.Context, stmt Statement, placeholderHints tree.PlaceholderTypes,
) (*PreparedStatement, error) {
	if placeholderHints == nil {
		placeholderHints = make(tree.PlaceholderTypes, stmt.NumPlaceholders)
	}

	prepared := &PreparedStatement{
		PrepareMetadata: sqlbase.PrepareMetadata{
			PlaceholderTypesInfo: tree.PlaceholderTypesInfo{
				TypeHints: placeholderHints,
			},
		},
		memAcc:   ex.sessionMon.MakeBoundAccount(),
		refCount: 1,
	}
	// NB: if we start caching the plan, we'll want to keep around the memory
	// account used for the plan, rather than clearing it.
	defer prepared.memAcc.Clear(ctx)

	if stmt.AST == nil {
		return prepared, nil
	}
	prepared.Statement = stmt.Statement

	// Point to the prepared state, which can be further populated during query
	// preparation.
	stmt.Prepared = prepared

	if err := tree.ProcessPlaceholderAnnotations(stmt.AST, placeholderHints); err != nil {
		return nil, err
	}
	// Preparing needs a transaction because it needs to retrieve db/table
	// descriptors for type checking.
	// TODO(andrei): Needing a transaction for preparing seems non-sensical, as
	// the prepared statement outlives the txn. I hope that it's not used for
	// anything other than getting a timestamp.
	txn := client.NewTxn(ctx, ex.server.cfg.DB, ex.server.cfg.NodeID.Get(), client.RootTxn)

	p := &ex.planner
	ex.resetPlanner(ctx, p, txn, ex.server.cfg.Clock.PhysicalTime() /* stmtTS */, stmt.NumAnnotations)
	p.stmt = &stmt
	flags, err := ex.populatePrepared(ctx, txn, placeholderHints, p)
	if err != nil {
		txn.CleanupOnError(ctx, err)
		return nil, err
	}
	if err := txn.CommitOrCleanup(ctx); err != nil {
		return nil, err
	}

	// Account for the memory used by this prepared statement.
	if err := prepared.memAcc.Grow(ctx, prepared.MemoryEstimate()); err != nil {
		return nil, err
	}
	ex.updateOptCounters(flags)
	return prepared, nil
}

// populatePrepared analyzes and type-checks the query and populates
// stmt.Prepared.
func (ex *connExecutor) populatePrepared(
	ctx context.Context, txn *client.Txn, placeholderHints tree.PlaceholderTypes, p *planner,
) (planFlags, error) {
	stmt := p.stmt
	if err := p.semaCtx.Placeholders.Init(stmt.NumPlaceholders, placeholderHints); err != nil {
		return 0, err
	}
	prepared := stmt.Prepared

	p.extendedEvalCtx.PrepareOnly = true

	protoTS, err := p.isAsOf(stmt.AST)
	if err != nil {
		return 0, err
	}
	if protoTS != nil {
		p.semaCtx.AsOfTimestamp = protoTS
		txn.SetFixedTimestamp(ctx, *protoTS)
	}

	// PREPARE has a limited subset of statements it can be run with. Postgres
	// only allows SELECT, INSERT, UPDATE, DELETE and VALUES statements to be
	// prepared.
	// See: https://www.postgresql.org/docs/current/static/sql-prepare.html
	// However, we allow a large number of additional statements.
	// As of right now, the optimizer only works on SELECT statements and will
	// fallback for all others, so this should be safe for the foreseeable
	// future.
	var flags planFlags
	var isCorrelated bool
	if optMode := ex.sessionData.OptimizerMode; optMode != sessiondata.OptimizerOff {
		log.VEvent(ctx, 2, "preparing using optimizer")
		var err error
		flags, isCorrelated, err = p.prepareUsingOptimizer(ctx)
		if err == nil {
			log.VEvent(ctx, 2, "optimizer prepare succeeded")
			// stmt.Prepared fields have been populated.
			return flags, nil
		}
		log.VEventf(ctx, 1, "optimizer prepare failed: %v", err)
		if !canFallbackFromOpt(err, optMode, stmt) {
			return 0, err
		}
		flags.Set(planFlagOptFallback)
		log.VEvent(ctx, 1, "prepare falls back on heuristic planner")
	} else {
		log.VEvent(ctx, 2, "optimizer disabled (prepare)")
	}

	// Fallback on the heuristic planner if the optimizer was not enabled or used:
	// create a plan for the statement to figure out the typing, then close the
	// plan.
	prepared.AnonymizedStr = anonymizeStmt(stmt.AST)
	if err := p.prepare(ctx, stmt.AST); err != nil {
		err = enhanceErrWithCorrelation(err, isCorrelated)
		return 0, err
	}

	if p.curPlan.plan == nil {
		// Statement with no result columns and no support for placeholders.
		//
		// Note: we're combining `flags` which comes from
		// `prepareUsingOptimizer`, with `p.curPlan.flags` which ensures
		// the new flags combine with the existing flags (this is used
		// e.g. to maintain the count of times the optimizer was used).
		return flags | p.curPlan.flags, nil
	}
	defer p.curPlan.close(ctx)

	prepared.Columns = p.curPlan.columns()
	for _, c := range prepared.Columns {
		if err := checkResultType(c.Typ); err != nil {
			return 0, err
		}
	}
	// Verify that all placeholder types have been set.
	if err := p.semaCtx.Placeholders.Types.AssertAllSet(); err != nil {
		return 0, err
	}
	prepared.Types = p.semaCtx.Placeholders.Types
	// The flags are combined, see the comment above for why.
	return flags | p.curPlan.flags, nil
}

func (ex *connExecutor) execBind(
	ctx context.Context, bindCmd BindStmt,
) (fsm.Event, fsm.EventPayload) {

	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		return eventNonRetriableErr{IsCommit: fsm.False}, eventNonRetriableErrPayload{err: err}
	}

	portalName := bindCmd.PortalName
	// The unnamed portal can be freely overwritten.
	if portalName != "" {
		if _, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]; ok {
			return retErr(pgerror.Newf(
				pgcode.DuplicateCursor, "portal %q already exists", portalName))
		}
	} else {
		// Deallocate the unnamed portal, if it exists.
		ex.deletePortal(ctx, "")
	}

	ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[bindCmd.PreparedStatementName]
	if !ok {
		return retErr(pgerror.Newf(
			pgcode.InvalidSQLStatementName,
			"unknown prepared statement %q", bindCmd.PreparedStatementName))
	}

	numQArgs := uint16(len(ps.InferredTypes))

	// Decode the arguments, except for internal queries for which we just verify
	// that the arguments match what's expected.
	qargs := make(tree.QueryArguments, numQArgs)
	if bindCmd.internalArgs != nil {
		if len(bindCmd.internalArgs) != int(numQArgs) {
			return retErr(
				pgwirebase.NewProtocolViolationErrorf(
					"expected %d arguments, got %d", numQArgs, len(bindCmd.internalArgs)))
		}
		for i, datum := range bindCmd.internalArgs {
			t := ps.InferredTypes[i]
			if oid := datum.ResolvedType().Oid(); datum != tree.DNull && oid != t {
				return retErr(
					pgwirebase.NewProtocolViolationErrorf(
						"for argument %d expected OID %d, got %d", i, t, oid))
			}
			qargs[i] = datum
		}
	} else {
		qArgFormatCodes := bindCmd.ArgFormatCodes

		// If a single code is specified, it is applied to all arguments.
		if len(qArgFormatCodes) != 1 && len(qArgFormatCodes) != int(numQArgs) {
			return retErr(pgwirebase.NewProtocolViolationErrorf(
				"wrong number of format codes specified: %d for %d arguments",
				len(qArgFormatCodes), numQArgs))
		}
		// If a single format code was specified, it applies to all the arguments.
		if len(qArgFormatCodes) == 1 {
			fmtCode := qArgFormatCodes[0]
			qArgFormatCodes = make([]pgwirebase.FormatCode, numQArgs)
			for i := range qArgFormatCodes {
				qArgFormatCodes[i] = fmtCode
			}
		}

		if len(bindCmd.Args) != int(numQArgs) {
			return retErr(
				pgwirebase.NewProtocolViolationErrorf(
					"expected %d arguments, got %d", numQArgs, len(bindCmd.Args)))
		}

		ptCtx := tree.NewParseTimeContext(ex.sessionData.DurationAdditionMode,
			ex.state.sqlTimestamp.In(ex.sessionData.DataConversion.Location))

		for i, arg := range bindCmd.Args {
			k := tree.PlaceholderIdx(i)
			t := ps.InferredTypes[i]
			if arg == nil {
				// nil indicates a NULL argument value.
				qargs[k] = tree.DNull
			} else {
				d, err := pgwirebase.DecodeOidDatum(ptCtx, t, qArgFormatCodes[i], arg)
				if err != nil {
					return retErr(pgerror.Wrapf(err, pgcode.ProtocolViolation,
						"error in argument for %s", k))
				}
				qargs[k] = d
			}
		}
	}

	numCols := len(ps.Columns)
	if (len(bindCmd.OutFormats) > 1) && (len(bindCmd.OutFormats) != numCols) {
		return retErr(pgwirebase.NewProtocolViolationErrorf(
			"expected 1 or %d for number of format codes, got %d",
			numCols, len(bindCmd.OutFormats)))
	}

	columnFormatCodes := bindCmd.OutFormats
	if len(bindCmd.OutFormats) == 1 {
		// Apply the format code to every column.
		columnFormatCodes = make([]pgwirebase.FormatCode, numCols)
		for i := 0; i < numCols; i++ {
			columnFormatCodes[i] = bindCmd.OutFormats[0]
		}
	}

	// Create the new PreparedPortal.
	if err := ex.addPortal(
		ctx, portalName, bindCmd.PreparedStatementName, ps, qargs, columnFormatCodes,
	); err != nil {
		return retErr(err)
	}

	if log.V(2) {
		log.Infof(ctx, "portal: %q for %q, args %q, formats %q",
			portalName, ps.Statement, qargs, columnFormatCodes)
	}

	return nil, nil
}

// addPortal creates a new PreparedPortal on the connExecutor.
//
// It is illegal to call this when a portal with that name already exists (even
// for anonymous portals).
func (ex *connExecutor) addPortal(
	ctx context.Context,
	portalName string,
	psName string,
	stmt *PreparedStatement,
	qargs tree.QueryArguments,
	outFormats []pgwirebase.FormatCode,
) error {
	if _, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]; ok {
		panic(fmt.Sprintf("portal already exists: %q", portalName))
	}

	portal, err := ex.newPreparedPortal(ctx, portalName, stmt, qargs, outFormats)
	if err != nil {
		return err
	}

	ex.extraTxnState.prepStmtsNamespace.portals[portalName] = portal
	return nil
}

func (ex *connExecutor) deletePreparedStmt(ctx context.Context, name string) {
	ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]
	if !ok {
		return
	}
	ps.decRef(ctx)
	delete(ex.extraTxnState.prepStmtsNamespace.prepStmts, name)
}

func (ex *connExecutor) deletePortal(ctx context.Context, name string) {
	portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[name]
	if !ok {
		return
	}
	portal.decRef(ctx)
	delete(ex.extraTxnState.prepStmtsNamespace.portals, name)
}

func (ex *connExecutor) execDelPrepStmt(
	ctx context.Context, delCmd DeletePreparedStmt,
) (fsm.Event, fsm.EventPayload) {
	switch delCmd.Type {
	case pgwirebase.PrepareStatement:
		_, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[delCmd.Name]
		if !ok {
			// The spec says "It is not an error to issue Close against a nonexistent
			// statement or portal name". See
			// https://www.postgresql.org/docs/current/static/protocol-flow.html.
			break
		}

		ex.deletePreparedStmt(ctx, delCmd.Name)
	case pgwirebase.PreparePortal:
		_, ok := ex.extraTxnState.prepStmtsNamespace.portals[delCmd.Name]
		if !ok {
			break
		}
		ex.deletePortal(ctx, delCmd.Name)
	default:
		panic(fmt.Sprintf("unknown del type: %s", delCmd.Type))
	}
	return nil, nil
}

func (ex *connExecutor) execDescribe(
	ctx context.Context, descCmd DescribeStmt, res DescribeResult,
) (fsm.Event, fsm.EventPayload) {

	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		return eventNonRetriableErr{IsCommit: fsm.False}, eventNonRetriableErrPayload{err: err}
	}

	switch descCmd.Type {
	case pgwirebase.PrepareStatement:
		ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[descCmd.Name]
		if !ok {
			return retErr(pgerror.Newf(
				pgcode.InvalidSQLStatementName,
				"unknown prepared statement %q", descCmd.Name))
		}

		res.SetInferredTypes(ps.InferredTypes)

		if stmtHasNoData(ps.AST) {
			res.SetNoDataRowDescription()
		} else {
			res.SetPrepStmtOutput(ctx, ps.Columns)
		}
	case pgwirebase.PreparePortal:
		portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[descCmd.Name]
		if !ok {
			return retErr(pgerror.Newf(
				pgcode.InvalidCursorName, "unknown portal %q", descCmd.Name))
		}

		if stmtHasNoData(portal.Stmt.AST) {
			res.SetNoDataRowDescription()
		} else {
			res.SetPortalOutput(ctx, portal.Stmt.Columns, portal.OutFormats)
		}
	default:
		return retErr(errors.AssertionFailedf(
			"unknown describe type: %s", errors.Safe(descCmd.Type)))
	}
	return nil, nil
}
