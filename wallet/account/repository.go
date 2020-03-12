package account

import (
	//"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"wallet/storage"
)

type Repository interface {
	Create(userId uuid.UUID) (Account, error)
	GetBalance(userId uuid.UUID) (float64, error)
	Deposit(userId uuid.UUID, amount uint) (float64, error)
	Withdraw(userId uuid.UUID, amount uint) (float64, error)
}

type repository struct {
	database *storage.Database
}

func NewRepository(db *storage.Database) Repository {
	return &repository{database: db}
}

func (r repository) Create(userId uuid.UUID) (Account, error) {
	// check if user has no account already
	var acc Account
	none := r.database.Where(Account{UserID: userId}).First(&acc).RecordNotFound()
	if !none {
		err := ErrUserHasAccount{userId: acc.UserID.String(), accountId: acc.ID.String()}
		return Account{}, err
	}

	// create account
	newAcc := zeroAccount(userId)
	result := r.database.Where(Account{UserID: userId}).Assign(newAcc).FirstOrCreate(&acc)
	if err := result.Error; err != nil {
		return Account{}, NewErrUnexpected(err)
	}

	return acc, nil
}

func (r repository) GetBalance(userId uuid.UUID) (float64, error) {
	acc, err := r.isAccAccessible(userId)
	if err != nil {
		return 0, err
	}

	return acc.Balance, nil
}

func (r repository) Deposit(userId uuid.UUID, amount uint) (float64, error) {
	acc, err := r.isAccAccessible(userId)
	if err != nil {
		return 0, err
	}

	// update balance with amount: add amount
	amtF := acc.Balance + float64(amount)
	result := r.database.Model(acc).Updates(Account{Balance: amtF})
	if err = result.Error; err != nil {
		return 0, NewErrUnexpected(err)
	}

	return acc.Balance, nil
}

func (r repository) Withdraw(userId uuid.UUID, amount uint) (float64, error) {
	acc, err := r.isAccAccessible(userId)
	if err != nil {
		return 0, err
	}

	// update balance with amount: subtract amount
	amtF := acc.Balance - float64(amount)
	result := r.database.Model(acc).Updates(Account{Balance: amtF})
	if err = result.Error; err != nil {
		return 0, NewErrUnexpected(err)
	}

	return acc.Balance, nil
}

func (r repository) isAccAccessible(userId uuid.UUID) (*Account, error) {
	var acc Account
	none := r.database.Where(Account{UserID: userId}).First(&acc).RecordNotFound()
	if none {
		err := ErrAccountAccess{reason: "Not Created. Report issue"}
		return nil, err
	}

	if acc.Status == StatusFrozen || acc.Status == StatusSuspended {
		return nil, ErrAccountAccess{reason: acc.Status}
	}

	return &acc, nil
}

func zeroAccount(userId uuid.UUID) *Account {
	return &Account{
		ID:              uuid.NewV4(),
		Balance:         0,
		Status:          StatusActive,
		AccountType:     TypeCurrent,
		UserID:          userId,
		CreationDate:    time.Now(),
		LastUpdatedDate: time.Now(),
	}
}