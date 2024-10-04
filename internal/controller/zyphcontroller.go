package controller

import zyp "github.com/Z3DRP/zportfolio-service/internal/zypher"

func CalculateZypher(txt string, shft, shftCount, hshCount int, alt, ignSpc, restcHsh bool) (string, error) {
	zypher := zyp.NewZypher(
		zyp.WithShift(shft),
		zyp.WithShiftIterCount(shftCount),
		zyp.WithHashIterCount(hshCount),
		zyp.WithAlternate(alt),
		zyp.WithIgnoreSpace(ignSpc),
		zyp.WithRestrictedHashShift(restcHsh),
	)

	result, err := zypher.Zyph(txt)
	if err != nil {
		return "", err
	}
	return result, nil
}
