package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
	// проще было бы сразу
	// randRange := rand.New(rand.NewSource(time.Now().UnixNano()))
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	// uniqueNumber := randRange.Intn(1000000) // при создании посылки в БД мы НЕ указываем айди, а ждем его от БД

	return Parcel{
		// Number:    uniqueNumber, // при создании посылки в БД мы НЕ указываем айди, а ждем его от БД
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// OpenConnection устанавливает соединение с базой данных
// Я вынес ее отдельно
func OpenConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := OpenConnection() // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	fmt.Println("Тестовая посылка:", parcel)
	id, err := store.Add(parcel) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	/*if err != nil {
		fmt.Println("Ошибка при добавлении посылки:", err)
		return
	}*/
	// fmt.Println("ID:", id) // ID: 0
	// fmt.Println("Error:", err) // Error: missing named argument "client"
	require.NoError(t, err)
	// require.NotEqual(t, 0, id)
	require.NotEmpty(t, id) // так побольше закроет ошибок
	parcel.Number = id

	// get
	storedParcel, err := store.Get(id) // получите только что добавленную посылку, убедитесь в отсутствии ошибки
	require.NoError(t, err)
	require.Equal(t, parcel, storedParcel) // проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	// delete
	err = store.Delete(id) // удалите добавленную посылку, убедитесь в отсутствии ошибки
	require.NoError(t, err)

	// check
	_, err = store.Get(id) // проверьте, что посылку больше нельзя получить из БД
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := OpenConnection() // настройте подключение к БД

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel()) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	require.NoError(t, err)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress) // обновите адрес, убедитесь в отсутствии ошибки
	require.NoError(t, err)

	// check
	parcel, err := store.Get(id) // получите добавленную посылку и убедитесь, что адрес обновился
	require.NoError(t, err)
	require.Equal(t, newAddress, parcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := OpenConnection() // настройте подключение к БД

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel()) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	require.NoError(t, err)

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus) // обновите статус, убедитесь в отсутствии ошибки
	require.NoError(t, err)

	// check
	parcel, err := store.Get(id) // получите добавленную посылку и убедитесь, что статус обновился
	require.NoError(t, err)
	require.Equal(t, newStatus, parcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := OpenConnection() // настройте подключение к БД

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	// parcelMap := map[int]Parcel{}
	parcelMap := make(map[int]Parcel) // так понятнее

	// задаём всем посылкам один и тот же идентификатор клиента
	// randRange.Int() // это бы позволило получать более случайные идентификаторы клиентов
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// var storedParcels []Parcel
	storedParcels, err := store.GetByClient(client)    // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)                            // убедитесь в отсутствии ошибки
	require.Equal(t, len(parcels), len(storedParcels)) // убедитесь, что количество полученных посылок совпадает с количеством добавленных
	// require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		require.Contains(t, parcelMap, parcel.Number)
		require.Equal(t, parcelMap[parcel.Number], parcel)
	}
}
