//go:build exclude
// +build exclude

/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package ladon_manager

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	gorp "gopkg.in/gorp.v1"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/runtime"
	sql2 "github.com/pydio/cells/v4/common/sql"

	. "github.com/ory/ladon"
)

// SQLManager is a postgres implementation for Manager to store policies persistently.
type SQLManager struct {
	db       *sqlx.DB
	database string
}

// NewSQLManager initializes a new SQLManager for given db instance.
func NewSQLManager(db *sqlx.DB, schema []string) *SQLManager {
	database := db.DriverName()
	switch database {
	case "pgx", "pq":
		database = "postgres"
	}

	return &SQLManager{
		db:       db,
		database: database,
	}
}

// MigrateMigrationTable checks if migration table exists. If not, we are upgrading
// from v3 and we need mimick the new one
func (s *SQLManager) MigrateMigrationTable(tableName string) error {
	if rows, er := s.db.Query("SELECT * FROM " + tableName); er == nil && rows.Next() {
		// Table exists, nothing to do
		return nil
	}
	if _, er := s.db.Query("SELECT * FROM gorp_migrations"); er != nil {
		// Table gorp_migration does not exist, ignore
		return nil
	}

	dbMap := &gorp.DbMap{Db: s.db.DB, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	dbMap.AddTableWithNameAndSchema(migrate.MigrationRecord{}, "", tableName).SetKeys(false, "Id")
	if er := dbMap.CreateTablesIfNotExists(); er != nil {
		return er
	}
	oldRows, er := s.db.Query("SELECT * from gorp_migrations WHERE id='1' OR id='2' OR id='3'")
	if er != nil {
		return er
	}
	for oldRows.Next() {
		mig := &migrate.MigrationRecord{}
		if er := oldRows.Scan(&mig.Id, &mig.AppliedAt); er != nil {
			continue
		}
		if er := dbMap.Insert(&migrate.MigrationRecord{
			Id:        mig.Id,
			AppliedAt: mig.AppliedAt,
		}); er != nil {
			return er
		}
	}
	if er := oldRows.Close(); er != nil {
		return er
	}
	res, er := s.db.Exec("DELETE from gorp_migrations WHERE id='1' OR id='2' OR id='3'")
	if er != nil {
		return er
	}
	del, _ := res.RowsAffected()
	if del > 0 {
		log.Logger(runtime.WithServiceName(context.Background(), "pydio.grpc.policy")).Info(fmt.Sprintf("Migrated %d rows from old gorp_migrations table to %s\n", del, tableName))
	}

	return nil
}

// CreateSchemas creates ladon_policy tables
func (s *SQLManager) CreateSchemas(schema, table string) (int, error) {
	if _, ok := Migrations[s.database]; !ok {
		return 0, errors.Errorf("Database %s is not supported", s.database)
	}

	source := Migrations[s.database].Migrations

	migrate.SetSchema(schema)
	migrate.SetTable(table)
	if er := s.MigrateMigrationTable(table); er != nil {
		return 0, errors.Wrapf(er, "Could not create migration table")
	}
	n, err := migrate.Exec(s.db.DB, s.database, source, migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema for %s, applied %d migrations", table, n)
	}
	return n, nil
}

// Update updates an existing policy.
func (s *SQLManager) Update(policy Policy) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(policy.GetID(), tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

// Create inserts a new policy
func (s *SQLManager) Create(policy Policy) (err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.create(policy, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLManager) create(policy Policy, tx *sqlx.Tx) (err error) {
	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if _, ok := Migrations[s.database]; !ok {
		return errors.Errorf("Database %s is not supported", s.database)
	}

	if _, err = tx.Exec(s.db.Rebind(Migrations[s.database].QueryInsertPolicy), policy.GetID(), policy.GetDescription(), policy.GetEffect(), conditions); err != nil {
		return errors.WithStack(err)
	}

	type relation struct {
		p []string
		t string
	}
	var relations = []relation{
		{p: policy.GetActions(), t: "action"},
		{p: policy.GetResources(), t: "resource"},
		{p: policy.GetSubjects(), t: "subject"},
	}

	for _, rel := range relations {
		var query string
		var queryRel string

		switch rel.t {
		case "action":
			query = Migrations[s.database].QueryInsertPolicyActions
			queryRel = Migrations[s.database].QueryInsertPolicyActionsRel
		case "resource":
			query = Migrations[s.database].QueryInsertPolicyResources
			queryRel = Migrations[s.database].QueryInsertPolicyResourcesRel
		case "subject":
			query = Migrations[s.database].QueryInsertPolicySubjects
			queryRel = Migrations[s.database].QueryInsertPolicySubjectsRel
		}

		for _, template := range rel.p {
			h := sha256.New()
			h.Write([]byte(template))
			id := fmt.Sprintf("%x", h.Sum(nil))

			compiled, err := compiler.CompileRegex(template, policy.GetStartDelimiter(), policy.GetEndDelimiter())
			if err != nil {
				return errors.WithStack(err)
			}

			if _, err := tx.Exec(s.db.Rebind(query), id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1); err != nil {
				return errors.WithStack(err)
			}
			if _, err := tx.Exec(s.db.Rebind(queryRel), policy.GetID(), id); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (s *SQLManager) FindRequestCandidates(r *Request) (Policies, error) {
	query := Migrations[s.database].QueryRequestCandidates

	ctx, cancel := context.WithTimeout(context.Background(), sql2.DefaultConnectionTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, s.db.Rebind(query), r.Subject, r.Subject)
	if err == sql.ErrNoRows {
		return nil, NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (s *SQLManager) FindPoliciesForResource(resource string) (Policies, error) {
	return nil, fmt.Errorf("FindPoliciesForResource not implemented inside common/sql/ladon-manager/SQLManager as Ladon Manager interface")
}

func (s *SQLManager) FindPoliciesForSubject(subject string) (Policies, error) {
	return nil, fmt.Errorf("FindPoliciesForSubject not implemented inside common/sql/ladon-manager/SQLManager as Ladon Manager interface")
}

func scanRows(rows *sql.Rows) (Policies, error) {
	var policies = map[string]*DefaultPolicy{}

	for rows.Next() {
		var p DefaultPolicy
		var conditions []byte
		var resource, subject, action sql.NullString
		p.Actions = []string{}
		p.Subjects = []string{}
		p.Resources = []string{}

		if err := rows.Scan(&p.ID, &p.Effect, &conditions, &p.Description, &subject, &resource, &action); err == sql.ErrNoRows {
			return nil, NewErrResourceNotFound(err)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		p.Conditions = Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return nil, errors.WithStack(err)
		}

		if c, ok := policies[p.ID]; ok {
			if action.Valid {
				policies[p.ID].Actions = append(c.Actions, action.String)
			}

			if subject.Valid {
				policies[p.ID].Subjects = append(c.Subjects, subject.String)
			}

			if resource.Valid {
				policies[p.ID].Resources = append(c.Resources, resource.String)
			}
		} else {
			if action.Valid {
				p.Actions = []string{action.String}
			}

			if subject.Valid {
				p.Subjects = []string{subject.String}
			}

			if resource.Valid {
				p.Resources = []string{resource.String}
			}

			policies[p.ID] = &p
		}
	}

	var result = make(Policies, len(policies))
	var count int
	for _, v := range policies {
		v.Actions = uniq(v.Actions)
		v.Resources = uniq(v.Resources)
		v.Subjects = uniq(v.Subjects)
		result[count] = v
		count++
	}

	return result, nil
}

var getQuery = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	ladon_policy as p

LEFT JOIN ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN ladon_action as action ON ra.action = action.id
LEFT JOIN ladon_resource as resource ON rr.resource = resource.id

WHERE p.id=?`

var getAllQuery = `SELECT
	p.id, p.effect, p.conditions, p.description,
	subject.template as subject, resource.template as resource, action.template as action
FROM
	(SELECT * from ladon_policy ORDER BY id LIMIT ? OFFSET ?) as p

LEFT JOIN ladon_policy_subject_rel as rs ON rs.policy = p.id
LEFT JOIN ladon_policy_action_rel as ra ON ra.policy = p.id
LEFT JOIN ladon_policy_resource_rel as rr ON rr.policy = p.id

LEFT JOIN ladon_subject as subject ON rs.subject = subject.id
LEFT JOIN ladon_action as action ON ra.action = action.id
LEFT JOIN ladon_resource as resource ON rr.resource = resource.id`

// GetAll returns all policies
func (s *SQLManager) GetAll(limit, offset int64) (Policies, error) {
	query := s.db.Rebind(getAllQuery)

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// Get retrieves a policy.
func (s *SQLManager) Get(id string) (Policy, error) {
	query := s.db.Rebind(getQuery)

	rows, err := s.db.Query(query, id)
	if err == sql.ErrNoRows {
		return nil, NewErrResourceNotFound(err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	policies, err := scanRows(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if len(policies) == 0 {
		return nil, NewErrResourceNotFound(sql.ErrNoRows)
	}

	return policies[0], nil
}

// Delete removes a policy.
func (s *SQLManager) Delete(id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.delete(id, tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrap(err, rollErr.Error())
		}
		return errors.WithStack(err)
	}

	return nil
}

// Delete removes a policy.
func (s *SQLManager) delete(id string, tx *sqlx.Tx) error {
	_, err := tx.Exec(s.db.Rebind("DELETE FROM ladon_policy WHERE id=?"), id)
	return errors.WithStack(err)
}

func uniq(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
