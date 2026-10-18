package pool_test

import (
	"github.com/KonstantinPavlov/metric-service/internal/pool"
	"testing"
)

// mockResetableStruct — фейковая структура для тестирования, реализующая метод Reset.
type mockResetableStruct struct {
	value string
	slice []int
	ptr   *int
}

// Reset приводит состояние объекта к начальным значениям.
func (m *mockResetableStruct) Reset() {
	if m == nil {
		return
	}
	m.value = ""
	m.slice = m.slice[:0]
	if m.ptr != nil {
		*m.ptr = 0
	}
}

func TestPool(t *testing.T) {
	// Инициализируем пул с функцией-аллокатором
	p := pool.New(func() *mockResetableStruct {
		return &mockResetableStruct{}
	})

	// Берем объект из пустого пула — должен создаться новый чистый объект
	item1 := p.Get()
	if item1 == nil {
		t.Fatal("Get() вернул nil, ожидался новый объект")
	}
	if item1.value != "" || len(item1.slice) != 0 || item1.ptr != nil {
		t.Errorf("Get() вернул неинициализированный объект: %+v", item1)
	}

	// Модифицируем объект
	num := 42
	item1.value = "dirty_data"
	item1.slice = append(item1.slice, 1, 2, 3)
	item1.ptr = &num

	// 4. Возвращаем объект обратно в пул.
	p.Put(item1)
	item2 := p.Get()

	// Проверяем, что нам вернулся ровно тот же самый объект из памяти (повторное использование)
	if item1 != item2 {
		t.Error("Пул вернул новый объект вместо повторного использования старого")
	}

	// Проверяем, что метод Put корректно очистил состояние объекта
	if item2.value != "" {
		t.Errorf("Поле value не очистилось: ожидалось '', получили '%s'", item2.value)
	}
	if len(item2.slice) != 0 {
		t.Errorf("Слайс не обрезался: ожидалась длина 0, получили %d", len(item2.slice))
	}
	if item2.ptr != nil && *item2.ptr != 0 {
		t.Errorf("Значение по указателю не сбросилось: ожидалось 0, получили %d", *item2.ptr)
	}
}

func TestPool_PutNil(t *testing.T) {
	p := pool.New(func() *mockResetableStruct {
		return &mockResetableStruct{}
	})

	// Проверяем защиту от паники при передаче nil в Put
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Пул запаниковал при передаче nil: %v", r)
		}
	}()

	p.Put(nil)
}
