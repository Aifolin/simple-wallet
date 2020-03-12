package account

import (
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"

	"wallet/data"
)

const (
	minimumDepositAmount    = 10
	minimumWithdrawalAmount = 0
)

type Interactor interface {
	GetBalance(userId uuid.UUID) (float64, error)
	Deposit(userId uuid.UUID, amount uint) (float64, error)
	Withdraw(userId uuid.UUID, amount uint) (float64, error)
}

func NewInteractor(repository Repository, usersChan data.ChanNewUsers) Interactor {
	intr := &interactor{repository: repository, usersChannel: usersChan}

	go intr.listenOnNewUsers()

	return intr
}

type interactor struct {
	repository Repository
	usersChannel data.ChanNewUsers
}

func (i interactor) CreateAccount(userId uuid.UUID) (Account, error) {
	acc, err := i.repository.Create(userId)
	if err != nil {
		return Account{}, err
	}
	return acc, nil
}

func (i interactor) GetBalance(userId uuid.UUID) (float64, error) {
	bal, err := i.repository.GetBalance(userId)
	if err != nil {
		return 0, err
	}
	return bal, nil
}

func (i interactor) Deposit(userId uuid.UUID, amount uint) (float64, error) {
	if amount < 10 {
		return 0, ErrAmountBelowMinimum{
			Message: fmt.Sprintf("cannot deposit amounts less than %v", minimumDepositAmount),
		}
	}

	return i.repository.Deposit(userId, amount)
}

func (i interactor) Withdraw(userId uuid.UUID, amount uint) (float64, error) {
	if amount < 10 {
		return 0, ErrAmountBelowMinimum{
			Message: fmt.Sprintf("cannot withdraw amounts less than %v", minimumWithdrawalAmount),
		}
	}

	// we can implement a double withdrawal check here. That will prevent a user from
	// withdrawing same amount twice within a stipulated time interval.

	return i.repository.Withdraw(userId, amount)
}

func (i interactor) listenOnNewUsers() {
	for {
		select{
		case user := <- i.usersChannel.Reader:
			acc, err := i.CreateAccount(user.UserID)
			if err != nil {
				// we need to log this error
				log.Printf("error happened while creating account %v", err)
				return
			}
			// we log the account details if created
			log.Printf("account with id %v has been created successfully for userID %v", acc.ID, user.UserID)
			return
		}
	}
}