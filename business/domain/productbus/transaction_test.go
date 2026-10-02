package productbus

import (
	"bytes"
	"context"
	"testing"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/delegate"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

func TestNewWithTxDoesNotRegisterDelegate(t *testing.T) {
	var callbackCalls int
	log := logger.NewWithEvents(&bytes.Buffer{}, logger.LevelInfo, "TEST", nil, logger.Events{
		Info: func(ctx context.Context, r logger.Record) {
			if r.Message == "action-userdeleted" {
				callbackCalls++
			}
		},
	})
	d := delegate.New(log)

	var extensionCalls int
	extension := func(bus ExtBusiness) ExtBusiness {
		extensionCalls++
		return &testExtension{ExtBusiness: bus}
	}

	bus := NewBusiness(log, testUserBusiness{}, d, testStorer{}, extension)
	for range 2 {
		var err error
		bus, err = bus.NewWithTx(testTransaction{})
		if err != nil {
			t.Fatalf("NewWithTx: %s", err)
		}
	}

	if extensionCalls != 3 {
		t.Fatalf("extension calls: got %d, want 3", extensionCalls)
	}

	if err := d.Call(t.Context(), userbus.ActionDeletedData(uuid.New())); err != nil {
		t.Fatalf("Call: %s", err)
	}

	if callbackCalls != 1 {
		t.Fatalf("callback calls: got %d, want 1", callbackCalls)
	}
}

type testStorer struct {
	Storer
}

func (testStorer) NewWithTx(sqldb.CommitRollbacker) (Storer, error) {
	return testStorer{}, nil
}

type testUserBusiness struct {
	userbus.ExtBusiness
}

func (testUserBusiness) NewWithTx(sqldb.CommitRollbacker) (userbus.ExtBusiness, error) {
	return testUserBusiness{}, nil
}

type testExtension struct {
	ExtBusiness
}

func (ext *testExtension) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	return ext.ExtBusiness.NewWithTx(tx)
}

type testTransaction struct{}

func (testTransaction) Commit() error {
	return nil
}

func (testTransaction) Rollback() error {
	return nil
}
