package zypher

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"regexp"
	"sync"
	"unicode"
)

type Zypher struct {
	Shift             int // number of runes/digits to shift, defaults to 3
	ShiftIterCount    int // number of iterations to be applied to value being zyphered, defaults to 3
	HashIterCount     int // number of iterations to be hashed, defaults to 3
	currentHashCount  int // number to be locked by mutex and keep track of current hash iterations performed
	hashCountMtx      *sync.Mutex
	Alternate         bool // if true when odd elements will reverse shift, defaults false
	IgnoreSpace       bool // if true will ignore space and leave them in, defaults to false
	RestrictHashShift bool // if true will only shift hash values within hahs digit range if false will shift digits outside hex range ie f could shift to j, default false
}

func DefaultZops() *Zypher {
	return &Zypher{
		Shift:             3,
		Alternate:         false,
		ShiftIterCount:    3,
		HashIterCount:     3,
		currentHashCount:  0,
		hashCountMtx:      &sync.Mutex{},
		IgnoreSpace:       false,
		RestrictHashShift: false,
	}
}

func NewZypher(ops ...func(*Zypher)) *Zypher {
	zy := DefaultZops()
	for _, op := range ops {
		op(zy)
	}
	return zy
}

func WithShift(shiftCount int) func(*Zypher) {
	return func(z *Zypher) {
		z.Shift = shiftCount
	}
}

func WithAlternate(wrp bool) func(*Zypher) {
	return func(z *Zypher) {
		z.Alternate = wrp
	}
}

func WithShiftIterCount(count int) func(*Zypher) {
	return func(z *Zypher) {
		z.ShiftIterCount = count
	}
}

func WithHashIterCount(count int) func(*Zypher) {
	return func(z *Zypher) {
		z.HashIterCount = count
	}
}

func WithIgnoreSpace(i bool) func(*Zypher) {
	return func(z *Zypher) {
		z.IgnoreSpace = i
	}
}

func WithRestrictedHashShift(i bool) func(*Zypher) {
	return func(z *Zypher) {
		z.RestrictHashShift = i
	}
}

func (z Zypher) AsciZyph(arg string) (string, error) {
	isValidString := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`).MatchString(arg)

	if !isValidString {
		InvalidString := errors.New("invalid string, only strings containing numbers, letters, and spaces are allowd")
		return "", InvalidString
	}

	if z.ShiftIterCount <= 0 {
		MissingShiftIterCount := errors.New(`invalid zypher, shift iter count expected but not found`)
		return "", MissingShiftIterCount
	}

	result := make([]rune, len(arg))
	var wg sync.WaitGroup

	for i := 0; i < z.ShiftIterCount; i++ {
		for indx, r := range arg {
			wg.Add(1)
			go func() {
				defer wg.Done()
				alt := false
				if z.Alternate {
					alt = indx%2 != 0
				}
				asciShift(&result[indx], r, z.Shift, alt, z.IgnoreSpace)
			}()
		}
		wg.Wait()
		arg = string(result)
	}
	//wg.Wait()
	return string(result), nil
}

func (z Zypher) HexZyph(arg string) (string, error) {
	isValidString := regexp.MustCompile(`^[a-fA-F0-9\s]+$`).MatchString(arg)
	if !isValidString {
		InvalidString := errors.New("invalid hex string, only A-F, a-f, and 0-9 allowed")
		return "", InvalidString
	}

	if z.ShiftIterCount <= 0 {
		MissingShiftIterCount := errors.New(`invalid zypher, shift iter count expected but not found`)
		return "", MissingShiftIterCount
	}
	// TODO HexZyph takes a hash then shifts it shifts for each shiftItercount
	result := make([]rune, len(arg))
	var wg sync.WaitGroup

	for i := 0; i < z.ShiftIterCount; i++ {
		for indx, r := range arg {
			wg.Add(1)
			go func() {
				defer wg.Done()
				alt := false
				if z.Alternate {
					alt = indx%2 != 0
				}
				if z.RestrictHashShift {
					hexShift(&result[indx], r, z.Shift, alt, z.IgnoreSpace)
				}
				if !z.RestrictHashShift {
					shift(&result[indx], r, z.Shift, alt, z.IgnoreSpace)
				}
			}()
		}
		wg.Wait()
		arg = string(result)
	}
	// wg.Wait()
	return string(result), nil
}

func (z Zypher) Zyph(arg string) (string, error) {
	isValidString := regexp.MustCompile(`^[a-zA-Z0-9!.,'"?_=\-+@#$%&()\s]+$`).MatchString(arg)
	if !isValidString {
		InvalidString := errors.New(`invalid string, only strings containing alpha numerics, spaces, and symbols ! . , ' " ? _ = - + @ # $ % & () are allowed`)
		return "", InvalidString
	}

	if z.ShiftIterCount <= 0 {
		MissingShiftIterCount := errors.New(`invalid zypher, shift iter count expected but not found`)
		return "", MissingShiftIterCount
	}

	if z.HashIterCount <= 0 {
		MissingHashIterCount := errors.New(`invalid zypher, hash iter count expected but not found`)
		return "", MissingHashIterCount
	}

	// TODO Zyph shifts string then hashes it foreach shiftItercount and foreach hashItercount
	result := make([]rune, len(arg))
	var wg sync.WaitGroup

	for i := 0; i < z.ShiftIterCount; i++ {
		for indx, r := range arg {
			wg.Add(1)
			go func() {
				defer wg.Done()
				alt := false
				if z.Alternate {
					alt = indx%2 != 0
				}

				shift(&result[indx], r, z.Shift, alt, z.IgnoreSpace)
			}()
		}
		wg.Wait()
		if z.currentHashCount <= z.HashIterCount {
			hsh := sha512.New()
			hsh.Write([]byte(string(result)))
			bs := hsh.Sum(nil)
			arg = hex.EncodeToString(bs)
			z.hashCountMtx.Lock()
			z.currentHashCount++
			z.hashCountMtx.Unlock()
		}
	}
	return string(result), nil
}

func (z Zypher) ZypHash(arg string) (string, error) {
	isValidString := regexp.MustCompile(`^[a-zA-Z0-9!.,'"?_=\-+@#$%&()\s]+$`).MatchString(arg)

	if !isValidString {
		InvalidString := errors.New(`invalid string, only strings containing alpha numerics, spaces, and symbols ! . , ' " ? _ = - + @ # $ % & () are allowed`)
		return " ", InvalidString
	}

	result := arg
	for i := 0; i < z.HashIterCount; i++ {
		hsh := sha512.New()
		// sidx := len(result) - 4
		// slt := result[sidx:]
		hsh.Write([]byte(result))
		bs := hsh.Sum(nil)
		result = hex.EncodeToString(bs)
	}

	return result, nil
}

func asciShift(target *rune, r rune, shf int, alt bool, ignSpc bool) {
	switch alt {
	case true:
		if shf > 0 {
			shf = -shf
		} else {
			shf = +shf
		}
	}

	if r >= '0' && r <= '9' {
		*target = '0' + (r-'0'+rune(shf)+10)%10
	} else if r >= 'A' && r <= 'Z' {
		*target = 'A' + (r-'A'+rune(shf)+26)%26

	} else if r >= 'a' && r <= 'z' {
		*target = 'a' + (r-'a'+rune(shf)+26)%26
	} else if r == ' ' {
		if !ignSpc {
			*target = 'x'
			// *target = 32 + (r-32+rune(shf)+95)%95
		} else {
			*target = ' '
		}
	} else {
		*target = r
	}
}

func hexShift(target *rune, r rune, shf int, alt, ignSpc bool) {
	switch alt {
	case true:
		if shf > 0 {
			shf = -shf
		} else {
			shf = +shf
		}
	}

	if r >= '0' && r <= '9' {
		*target = '0' + (r-'0'+rune(shf)+10)%10
	} else if r >= 'a' && r <= 'f' {
		*target = 'a' + (r-'a'+rune(shf)+6)%6
	} else if r >= 'A' && r <= 'F' {
		*target = 'A' + (r-'A'+rune(shf)+6)%6
	} else if r >= 32 && r <= 126 {
		if r == 32 && ignSpc {
			*target = r
		} else {
			*target = 32 + (r-32+rune(shf)+95)%95
		}
	}
}

func shift(target *rune, r rune, shf int, alt, ignSpc bool) {
	switch alt {
	case true:
		if shf > 0 {
			shf = -shf
		} else {
			shf = +shf
		}
	}

	if unicode.IsDigit(r) {
		*target = '0' + (r-'0'+rune(shf)+10)%10
	} else if unicode.IsUpper(r) {
		*target = 'A' + (r-'A'+rune(shf)+26)%26
	} else if unicode.IsLower(r) {
		*target = 'a' + (r-'a'+rune(shf)+26)%26
	} else if r >= 32 && r <= 126 {
		if r == 32 && ignSpc {
			*target = r
		} else {
			*target = 32 + (r-32+rune(shf)+95)%95
		}
	}
}
