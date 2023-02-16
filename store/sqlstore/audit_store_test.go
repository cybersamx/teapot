package sqlstore

import (
	"context"
	"testing"

	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	newAuditStoreFns = map[string]newTestStoreFunc{
		"mysql":    newAuditMySQLStore,
		"postgres": newAuditPostgresStore,
		"sqlite":   newAuditSQLiteStore,
	}
)

var (
	testAudits = []*model.Audit{
		{
			RequestID:     "72b24bea-ffd5-4914-9a29-b01408ec9a7a",
			ClientAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9",
			ClientAddress: "40.59.137.210",
			StatusCode:    201,
			Event:         "createUser",
		},
		{
			RequestID:     "5efd4510-ad43-11ed-afa1-0242ac120002",
			ClientAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
			ClientAddress: "40.59.137.212",
			StatusCode:    200,
			Event:         "getUser",
		},
		{
			RequestID:     "f7f52385-bc81-4db8-97e9-f362c7f00cf6",
			ClientAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
			ClientAddress: "40.59.137.212",
			StatusCode:    401,
			Event:         "getUser",
			Error:         "jwt expired",
		},
	}
)

func fillAudits(t *testing.T, ctx context.Context, as store.AuditStore, audits []*model.Audit) []*model.Audit {
	err := as.Clear(ctx)
	require.NoError(t, err)

	saves, err := common.FillStore[*model.Audit](ctx, as, audits)
	require.NoError(t, err)
	require.Len(t, saves, len(audits))

	return saves
}

func testAuditSQLStoreGet(t *testing.T, newTestStoreFn newTestStoreFunc) {
	sqlstore := newTestStoreFn(t)

	ctx := newTestContext(t)
	us := sqlstore.Audits()
	fillAudits(t, ctx, us, testAudits)

	// Get the first audit (happy path).
	user, err := us.Get(ctx, "72b24bea-ffd5-4914-9a29-b01408ec9a7a")
	assert.NoError(t, err)
	diff := pretty.Compare(testAudits[0], user)
	assert.Emptyf(t, diff, "want: %+v, got: %+v", testAudits[0], user)

	// Get non-existent user.
	user, err = us.Get(ctx, "non-exist-uuid")
	assert.ErrorIs(t, err, store.ErrNoRows)
	assert.Nil(t, user)
}

func TestAuditSQLStore_Get(t *testing.T) {
	for platform, fn := range newAuditStoreFns {
		t.Run("With "+platform, func(t *testing.T) {
			testAuditSQLStoreGet(t, fn)
		})
	}
}

func testAuditSQLStoreInsert(t *testing.T, newTestStoreFn newTestStoreFunc) {
	sqlstore := newTestStoreFn(t)

	ctx := newTestContext(t)
	as := sqlstore.Audits()
	fillAudits(t, ctx, as, testAudits)

	// Audits for testing insert.
	auditWithID := model.Audit{
		RequestID:     "27e2fb56-1d15-11ed-861d-0242ac120002",
		ClientAgent:   "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:15.0) Gecko/20100101 Firefox/15.0.1",
		ClientAddress: "173.46.68.3",
		StatusCode:    404,
		Error:         "file1 not found",
		Event:         "getFile",
	}
	auditWithoutID := model.Audit{ // Insert will generate an id
		ClientAgent:   "curl/7.79.1",
		ClientAddress: "173.46.68.3",
		StatusCode:    404,
		Error:         "file1 not found",
		Event:         "getFile",
	}
	auditWithExistID := model.Audit{
		RequestID:     testAudits[0].RequestID,
		ClientAgent:   "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:15.0) Gecko/20100101 Firefox/15.0.1",
		ClientAddress: "173.46.68.3",
		StatusCode:    404,
		Error:         "file1 not found",
		Event:         "getFile",
	}

	tests := []struct {
		description string
		audit       *model.Audit
		wantErr     error
	}{
		{"Audit with unique id", &auditWithID, nil},
		{"Audit without id", &auditWithoutID, nil},
		{"Audit with existing id", &auditWithExistID, ErrSQLDuplicate},
	}
	
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			save, err := as.Insert(ctx, tc.audit)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, save)
			} else {
				assert.NoError(t, err)
				diff := pretty.Compare(tc.audit, save)
				assert.Emptyf(t, diff, "want: %+v, got: %+v", tc.audit, save)
			}
		})
	}
}

func TestUserSQLStore_Insert(t *testing.T) {
	for platform, fn := range newAuditStoreFns {
		t.Run("With "+platform, func(t *testing.T) {
			testAuditSQLStoreInsert(t, fn)
		})
	}
}
