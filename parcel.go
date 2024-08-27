// тут мы работаем с конкретной посылкой
package main

import (
	"database/sql"
)

// Итак, поле db типа *sql.DB - это указатель на соединение с базой данных.
type ParcelStore struct {
	db *sql.DB
}

/*
// метод Clear для очистки таблицы
func (s ParcelStore) Clear() error {
	stmt, err := s.db.Prepare("DELETE FROM parcel;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}
*/
// Передаем хранилище посылок в качестве аргумента функциям Add() и Get().
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// база данных автоматически генерирует уникальный номер для новой записи и возвращает его
func (s ParcelStore) Add(p Parcel) (int, error) {
	stmt, err := s.db.Prepare("INSERT INTO parcel (number, client, status, address, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	// мы получаем этот сгенерированный номер (см выше), используя result.LastInsertId() после вызова stmt.Exec()

	// моя первая версия не прошла тесты с ошибкой: UNIQUE constraint failed: parcel.number
	// по идее, ошибка связана с нарушением уникального ограничения на поле number в таблице parcel. Видимо, пытаюсь добавить новую запись с номером, который уже существует в таблице. Надо генерировать уникальный номер для каждой тестовой посылки, либо удалять предыдущую запись перед добавлением новой. Не понимаю, почему не очищает УИН...
	// UPD: Сделал метод Clear, не помогло...
	result, err := stmt.Exec(p.Number, p.Client, p.Status, p.Address, p.CreatedAt)

	// так не сработало: missing named argument "client"
	/*result, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)", p.Client, p.Status, p.Address, p.CreatedAt)
	 */
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	// return 0, nil // не понял, к чему это, когда вернуть нужно id
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

	// добавил defer rows.Close() для того, чтобы закрыть курсор (объект rows) после завершения работы с ним
	defer rows.Close()

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

	// Проверяем на наличие ошибок после цикла
	if err = rows.Err(); err != nil {
		return nil, err
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
