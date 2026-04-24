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
        sql.Named ("number", number),
    )
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
	// возвращаем только одну строку
		return Parcel{},err
	}
	return p, nil
}

// Получение всех посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// чтение строк из таблицы parcel по заданному client
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
        sql.Named ("client", client),
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
		sql.Named ("status", status),
		sql.Named ("number", number),
		)
		return err
}

// Обновление статуса при registered
func (s ParcelStore) SetAddress(number int, address string) error {
	// получаем посылку и её номер
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	
	//  проверка: если статус другой, возвращаем ошибку
	if p.Status != ParcelStatusRegistered {
		return errors.New("менять адрес можно только если значение статуса registered")
	}
	
	// если условие соблюдено, то выполняем запрос UPDATE, изменяем адрес у посылки
	_, err = s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
	    sql.Named("address", address),
		sql.Named("number", number),
	)
	return err
}

// Удаление строки при registered
func (s ParcelStore) Delete(number int) error {
	// получаем посылку и её номер
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	// проверка: если статус другой, возвращаем ошибку
	if p.Status != ParcelStatusRegistered {
		return errors.New("удалять строку можно только если значение статуса registered")
	}

	// если условие соблюдено, то выполняем запрос Delete, удаляем строку
	_, err = s.db.Exec("DELETE FROM parcel WHERE number = :number",
		sql.Named("number", number),
	)
	return err
}