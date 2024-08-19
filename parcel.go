// тут мы работаем с конкретной посылкой
package main

import (
	"database/sql"
)

// Итак, поле db типа *sql.DB - это указатель на соединение с базой данных.
type ParcelStore struct {
	db *sql.DB
}

// Передаем хранилище посылок в качестве аргумента функциям Add() и Get().
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	stmt, err := s.db.Prepare("INSERT INTO parcel (number, client, status, address, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(p.Number, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	// return 0, nil - не понял, что это
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	stmt, err := s.db.Prepare("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?")
	if err != nil {
		return Parcel{}, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(number)
	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err = row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	stmt, err := s.db.Prepare("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(client)
	if err != nil {
		return nil, err
	}

	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	stmt, err := s.db.Prepare("UPDATE parcel SET status = ? WHERE number = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(status, number)
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	stmt, err := s.db.Prepare("UPDATE parcel SET address = ? WHERE number = ? AND status = 'registered'")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(address, number)
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	stmt, err := s.db.Prepare("DELETE FROM parcel WHERE number = ? AND status = 'registered'")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(number)
	if err != nil {
		return err
	}

	return nil
}
