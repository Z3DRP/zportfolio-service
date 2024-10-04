package models

import "fmt"

type Responser interface {
	PrintRes() string
}

type PortoflioResponse struct {
	AboutDetails           Detail
	ProfessionalExperience []Experience
	Skills                 []Skill
}

func (pr PortoflioResponse) PrintRes() string {
	return fmt.Sprintf("About: %v; ProfessionalExperience: %v; Skills: %v", pr.AboutDetails, pr.ProfessionalExperience, pr.Skills)
}

func NewPortfolioResponse(d Detail, e []Experience, s []Skill) *PortoflioResponse {
	return &PortoflioResponse{
		AboutDetails:           d,
		ProfessionalExperience: e,
		Skills:                 s,
	}
}
