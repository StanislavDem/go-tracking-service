package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Добавление новой посылки
func (s ParcelStore) Add(p Parcel) (int, error) {
	// добавление строки в таблицу parcel, используя данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)
	if err != nil { // проверка, что INSERT прошёл успешно
		return 0, err
	}
	// возвращаем идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil { // проверка, что удалось получить id вставленной строки
		return 0, err
	}
	return int(id), err
}

// Получение посылки по номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	// чтение строки по заданному number
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number),
	)
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		// возвращаем только одну строку
		return Parcel{}, err
	}
	return p, nil
}

// Получение всех посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// чтение строк из таблицы parcel по заданному client
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client),
	)
	if err != nil { // запрос к базе db.Query
		return nil, err
	}
	defer rows.Close() // закрытие соединения после выхода из ф-ии

	var res []Parcel //сюда будем складывать все найденные посылки

	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	// возвращаем список посылок клиента
	return res, nil
}

// Обновление статуса
func (s ParcelStore) SetStatus(number int, status string) error {
	// обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number),
	)
	return err
}

// Обновление статуса при registered
func (s ParcelStore) SetAddress(number int, address string) error {
	// выполняем запрос UPDATE, изменяем адрес у посылки если status registered
	res, err := s.db.Exec(
		"UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)

	//  проверка: если проблема при подключении к db или не синтаксис SQL не верный, то возвращаем ошибку
	if err != nil {
		return err
	}

	// проверка: сколько строк было изменено, если ни одна строка не обновилось значит status не registered
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// проверка: если не изменилось ни одной строки, возвращаем ошибку
	if rowsAffected == 0 {
		return errors.New("менять адрес можно только если значение статуса registered")
	}
	return nil
}

// Удаление строки при registered
func (s ParcelStore) Delete(number int) error {
	// выполняем запрос DELETE, удаляем строку из таблицы parcel, если number совпадает и status = registered
	res, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)

	//  проверка: если проблема при подключении к db или не синтаксис SQL не верный, то возвращаем ошибку
	if err != nil {
		return err
	}

	// проверка: сколько строк было удалено, если ни одна строка не под условие, то rowsAffected = 0
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// проверка: если не удалено ни одной строки, то status не registered, возвращаем ошибку
	if rowsAffected == 0 {
		return errors.New("удалять строку можно только если значение статуса registered")
	}
	return nil
}
