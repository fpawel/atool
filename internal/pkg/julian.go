package pkg

import "time"

//calculate the Julian  date, provided it's within 209 years of Jan 2, 2006.
func TimeToJulian(t time.Time) float64 {
	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	a := (14 - month) / 12

	y := year + 4800 - a

	m := month + 12*a - 3

	julianDay := day + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045

	hour := float64(t.Hour())
	minute := float64(t.Minute())
	second := float64(t.Second()) + float64(t.Nanosecond())/1e9

	jf := (hour-12.)/24. + minute/1440. + second/86400.

	return float64(julianDay) + jf
}

func JulianToTime(julianDay float64) time.Time {
	jdn, jdf := intDec(julianDay)
	year, month, day := julianDateToGregorian(int(jdn))
	hour, hf := intDec(jdf * 24)
	minute, mf := intDec(hf * 60)
	sec, sf := intDec(mf * 60)
	ns, _ := intDec(sf * 1_000_000_000)
	return time.Date(year, month, day, int(hour)+12, int(minute), int(sec), int(ns), time.Local)
}

func intDec(v float64) (Int, Dec float64) {
	Int = float64(int64(v))
	Dec = v - Int
	return
}

func julianDateToGregorian(jdn int) (year int, month time.Month, day int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153

	day = e - (153*m+2)/5 + 1
	month = time.Month(m + 3 - 12*(m/10))
	year = 100*b + d - 4800 + m/10

	return
}
