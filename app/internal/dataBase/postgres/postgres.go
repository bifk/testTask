package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bifk/testTask/internal/domain/models"
	"github.com/bifk/testTask/internal/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"os"
	"time"
)

type DataBase struct {
	db *sql.DB
}

// Подключение базы данных и, при необходимости, первоначальные миграции
func New() (*DataBase, error) {
	const op = "dataBase.postgres.New"
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
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
	    balance DECIMAL CHECK (balance >= 0)
	);

`)
	defer stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err = db.Prepare(`
CREATE TABLE IF NOT EXISTS transaction(
	    id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	    amount DECIMAL,
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

// Создание кошельков при первом запуске
func (db *DataBase) Init(logg *logger.Logger) error {
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
			_, err = stmt.Exec(address, 100)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("%s: %w", op, err)
			}
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		logg.Info("Создалось 10 кошельков")

	}
	return nil
}

// Закрытие соединения с базой данных
func (db *DataBase) Stop() error {
	const op = "dataBase.postgres.Stop"
	err := db.db.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Метод полученения информации о балансе кошелька
func (db *DataBase) GetWallet(address string) (models.Wallet, error) {
	const op = "dataBase.postgres.GetWallet"

	stmt, err := db.db.Prepare("SELECT * FROM wallet WHERE address = $1")
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	var wallet models.Wallet
	err = stmt.QueryRow(address).Scan(&wallet.Address, &wallet.Balance)
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}

	return wallet, nil
}

// Метод отправки средств между кошельками
func (db *DataBase) Send(from string, to string, amount decimal.Decimal) error {
	const op = "dataBase.postgres.Send"

	// Создание транзакции
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	//Изменение кошелька-отправителя
	stmt, err := tx.Prepare("UPDATE wallet SET balance = balance - $2 WHERE address = $1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(from, amount)
	if err != nil {
		tx.Rollback()

		// Код ошибки ограничения базы данных
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" {
			return fmt.Errorf("Недостаточно средств на кошельке-отправителе")
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("Адрес кошелька-отправителя не найден")
	}

	//Изменение кошелька-получателя
	stmt, err = tx.Prepare("UPDATE wallet SET balance = balance + $2 WHERE address = $1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	res, err = stmt.Exec(to, amount)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("Адрес кошелька-получателя не найден")
	}

	//Создание транзакции
	stmt, err = tx.Prepare(`
		INSERT INTO transaction(from_address, to_address, transaction_time, amount) 
		VALUES ($1, $2, $3, $4)
`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(from, to, time.Now(), amount)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Метод возвращающий информацию о n последних переводах
func (db *DataBase) GetLast(count int) ([]models.Transaction, error) {
	const op = "dataBase.postgres.GetLast"

	var transactions []models.Transaction

	stmt, err := db.db.Prepare(`
		SELECT amount, from_address, to_address, transaction_time FROM transaction
	    ORDER BY transaction_time DESC LIMIT $1
	`)
	if err != nil {
		return []models.Transaction{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(count)
	if err != nil {
		return []models.Transaction{}, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(&transaction.Amount, &transaction.From, &transaction.To, &transaction.Time); err != nil {
			return []models.Transaction{}, fmt.Errorf("%s: %w", op, err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
