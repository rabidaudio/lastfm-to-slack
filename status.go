package main

const MaxLength = 100
const Sep = " - "
const Heart = " :heart:"
const Tail = "..."
const Min = 9

func GenerateStatus(title, artist, album string, loved bool) string {
	first := artist
	second := title
	if *AlbumMode {
		second = album
	}
	first = strip(first)
	second = strip(second)
	available := MaxLength - len(Sep)
	if loved {
		available -= len(Heart)
	}
	if len(first)+len(second) > available {
		// first try triming the second and see if it will fit
		if len(first)+Min <= available {
			second = truncate(second, available-len(first))
		} else {
			// we're going to need to trim both
			lf := len(first) * available / (len(first) + len(second))
			ls := available - lf
			// but we need to make sure if were going to truncate both
			// that each is at least Min characters
			if lf < Min {
				lf = Min
				if len(first) < Min {
					lf = len(first)
				}
				ls = available - lf
			} else if ls < Min {
				ls = Min
				if len(second) < Min {
					ls = len(second)
				}
				lf = available - ls
			}
			first = truncate(first, lf)
			second = truncate(second, ls)
		}
	}
	msg := first + Sep + second
	if loved {
		msg = msg + Heart
	}
	return msg
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	base := value[0 : limit-len(Tail)]
	return strip(base) + Tail
}

func strip(value string) string {
	for value[0] == ' ' {
		value = value[1:]
	}
	for value[len(value)-1] == ' ' {
		value = value[:len(value)-1]
	}
	return value
}
