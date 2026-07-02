package civ6save

import (
	"fmt"
	"hash/crc32"
)

// A religion's Symbol is the Civ6 CRC (^crc32.IEEE) of its RELIGION_* type
// string. That type — not the player-chosen custom name — determines which
// symbol the game draws, so it maps directly onto our Religion_<key>.png icons.
func religionTypeCRC(s string) uint32 { return ^crc32.ChecksumIEEE([]byte(s)) }

var religionIconByCRC = func() map[uint32]string {
	m := map[uint32]string{}
	add := func(typ, key string) { m[religionTypeCRC(typ)] = key }

	add("RELIGION_BUDDHISM", "Buddhism")
	add("RELIGION_CATHOLICISM", "Catholicism")
	add("RELIGION_CONFUCIANISM", "Confucianism")
	add("RELIGION_HINDUISM", "Hinduism")
	add("RELIGION_ISLAM", "Islam")
	add("RELIGION_JUDAISM", "Judaism")
	add("RELIGION_EASTERN_ORTHODOXY", "Orthodoxy")
	add("RELIGION_ORTHODOXY", "Orthodoxy")
	add("RELIGION_PROTESTANTISM", "Protestantism")
	add("RELIGION_SHINTO", "Shinto")
	add("RELIGION_SIKHISM", "Sikhism")
	add("RELIGION_TAOISM", "Taoism")
	add("RELIGION_ZOROASTRIANISM", "Zoroastrianism")
	add("RELIGION_PANTHEON", "Pantheon")
	for i := 1; i <= 12; i++ {
		add(fmt.Sprintf("RELIGION_CUSTOM_%d", i), fmt.Sprintf("Custom%d", i))
	}
	return m
}()

// ReligionIconKey maps a religion symbol to a stable icon key matching the
// Religion_<key>.png assets (e.g. "Custom10", "Islam"), or "" if unknown.
func ReligionIconKey(symbol uint32) string {
	return religionIconByCRC[symbol]
}
