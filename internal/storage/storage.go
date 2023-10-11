package storage

// мапа для хранения ссылок
var Pairs = make(map[string]string)

// FindByValue Поиск ключа по значению пары
func FindByValue(value string) (key string, ok bool) {
	for k, v := range Pairs {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}
