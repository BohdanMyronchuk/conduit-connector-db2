// Copyright © 2022 Meroxa, Inc & Yalantis.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iterator

import (
	"fmt"
	"strings"
)

const (
	queryIfExistTable = `
	SELECT count(*) AS count FROM  SysCat.Tables WHERE TabName='%s';
`
	queryCreateTable = `
		CREATE TABLE %s AS (
			SELECT *
			FROM %s
		) WITH NO DATA
	`
	queryAddColumns = `
	ALTER TABLE %s 
		ADD COLUMN %s varchar(10)
		ADD COLUMN %s TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		ADD COLUMN %s INT 
	`
	querySetNotNull = `
	ALTER TABLE %s ALTER COLUMN %s SET NOT NULL
	`
	querySetGeneratedIdentity = `
	ALTER TABLE %s ALTER COLUMN %s SET GENERATED BY DEFAULT AS IDENTITY (CYCLE)
	`
	queryReorgTable = `
	CALL sysproc.admin_cmd('reorg table %s')
	`
	queryAddIndex = `
	CREATE UNIQUE INDEX CONDUIT_TRACKING_%s_ID_UNIQUE_IND
	ON %s(%s)
	`
	queryTriggerTemplate = `
      CREATE OR REPLACE TRIGGER CONDUIT_TRIGGER_{{operation_type}}_{{table}}
      AFTER {{operation_type}} ON {{table}}
      REFERENCING {{row_type}} ROW AS rw
      FOR EACH ROW
      BEGIN ATOMIC
        INSERT INTO %s (%s) VALUES (%s,'{{operation_type}}');
      END
	`

	queryGetMaxValue = `SELECT max(%s) FROM %s`

	placeholderOperationType = "{{operation_type}}"
	placeholderTable         = "{{table}}"
	placeholderRowType       = "{{row_type}}"
)

type queryTriggers struct {
	queryTriggerCatchInsert string
	queryTriggerCatchUpdate string
	queryTriggerCatchDelete string
}

func buildTriggers(trackingTable, table string, columnsTypes map[string]string) queryTriggers {
	columnNames := make([]string, 0)

	for key := range columnsTypes {
		columnNames = append(columnNames, key)
	}

	nwValues := make([]string, len(columnNames))
	for i := range columnNames {
		nwValues[i] = fmt.Sprintf("rw.%s", columnNames[i])
	}

	columnNames = append(columnNames, columnOperationType)

	triggerTemplate := fmt.Sprintf(queryTriggerTemplate, trackingTable,
		strings.Join(columnNames, ","), strings.Join(nwValues, ","))

	triggerTemplate = strings.ReplaceAll(triggerTemplate, placeholderTable, table)

	queryTriggerInsert := strings.ReplaceAll(triggerTemplate, placeholderOperationType, string(ActionInsert))
	queryTriggerInsert = strings.ReplaceAll(queryTriggerInsert, placeholderRowType, "NEW")

	queryTriggerUpdate := strings.ReplaceAll(triggerTemplate, placeholderOperationType, string(ActionUpdate))
	queryTriggerUpdate = strings.ReplaceAll(queryTriggerUpdate, placeholderRowType, "NEW")

	queryTriggerDelete := strings.ReplaceAll(triggerTemplate, placeholderOperationType, string(ActionDelete))
	queryTriggerDelete = strings.ReplaceAll(queryTriggerDelete, placeholderRowType, "OLD")

	return queryTriggers{
		queryTriggerCatchInsert: queryTriggerInsert,
		queryTriggerCatchUpdate: queryTriggerUpdate,
		queryTriggerCatchDelete: queryTriggerDelete,
	}
}
