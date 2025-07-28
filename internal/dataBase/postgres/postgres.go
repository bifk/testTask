package postgres

import (
	"database/sql"
	"fmt"
	"github.com/bifk/testTask/internal/domain/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type DataBase struct {
	db *sql.DB
}

func New(connectionString string) (*DataBase, error) {
	const op = "dataBase.postgres.New"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS wallet(
	    address VARCHAR PRIMARY KEY,
	    amount DECIMAL
	);

`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err = db.Prepare(`
CREATE TABLE IF NOT EXISTS transaction(
	    id INTEGER PRIMARY KEY,
	    from_address VARCHAR,
	    to_address VARCHAR,
	    transaction_time TIMESTAMP
	);
`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &DataBase{db: db}, nil
}

func (db *DataBase) Init() error {
	const op = "dataBase.postgres.Init"

	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM wallet").Scan(&count)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if count == 0 {

		tx, err := db.db.Begin()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		stmt, err := tx.Prepare("INSERT INTO wallet VALUES ($1, $2)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: %w", op, err)
		}

		defer stmt.Close()

		for range 10 {
			address, _ := uuid.NewUUID()
			_, err = stmt.Exec(address, 1000)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("%s: %w", op, err)
			}
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func (db *DataBase) GetWallet(address string) (models.Wallet, error) {
	const op = "dataBase.postgres.GetWallet"

	stmt, err := db.db.Prepare("SELECT * FROM wallet WHERE address = ?")
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}
	var wallet models.Wallet
	err = stmt.QueryRow(address).Scan(&wallet.Address, &wallet.Amount)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}

	return wallet, nil
}

func (db *DataBase) UpdateWallet(address string, amount decimal.Decimal) (models.Wallet, error) {
	const op = "dataBase.postgres.UpdateWallet"

	stmt, err := db.db.Prepare("UPDATE wallet SET amount = ? WHERE address = ?")
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(amount, address)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}
	return models.Wallet{Address: address, Amount: amount}, nil
}
