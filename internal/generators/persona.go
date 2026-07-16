// Package generators creates fake applicant data for CVs.
package generators

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// Persona represents a generated fake applicant.
type Persona struct {
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Experience []Job    `json:"experience"`
	Education  []School `json:"education"`
	Skills     []string `json:"skills"`
}

// Job represents work experience.
type Job struct {
	Company  string `json:"company"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
	Desc     string `json:"desc"`
}

// School represents education.
type School struct {
	Name   string `json:"name"`
	Degree string `json:"degree"`
	Year   string `json:"year"`
}

// PersonaGenerator creates fake applicant personas.
type PersonaGenerator struct {
	firstNames []string
	lastNames  []string
	domains    []string
	companies  []string
	titles     []string
	skills     []string
	cities     []string
	schools    []string
	degrees    []string
}

// NewPersonaGenerator creates a persona generator with default data.
func NewPersonaGenerator() *PersonaGenerator {
	return &PersonaGenerator{
		firstNames: []string{
			"Alex", "Jordan", "Taylor", "Morgan", "Casey", "Riley", "Quinn",
			"Avery", "Jamie", "Drew", "Cameron", "Blake", "Parker", "Reese",
		},
		lastNames: []string{
			"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
			"Davis", "Martinez", "Anderson", "Taylor", "Thomas", "Jackson", "Lee",
		},
		domains: []string{
			"gmail.com", "outlook.com", "yahoo.com", "proton.me", "icloud.com",
		},
		companies: []string{
			"TechCorp", "DataSystems", "CloudNet", "InnovateLabs", "CodeWorks",
			"DigitalFirst", "AgileTeam", "DevOps Inc", "StartupHub", "ByteForce",
		},
		titles: []string{
			"Software Engineer", "Senior Developer", "Full Stack Developer",
			"Backend Engineer", "DevOps Engineer", "Data Engineer",
			"Platform Engineer", "Site Reliability Engineer", "Tech Lead",
		},
		skills: []string{
			"Go", "Python", "JavaScript", "TypeScript", "Rust", "Java",
			"Kubernetes", "Docker", "AWS", "GCP", "Azure", "Terraform",
			"PostgreSQL", "Redis", "MongoDB", "Kafka", "GraphQL", "gRPC",
			"React", "Node.js", "Linux", "Git", "CI/CD", "Agile",
		},
		cities: []string{
			"San Francisco, CA", "New York, NY", "Seattle, WA", "Austin, TX",
			"Boston, MA", "Denver, CO", "Portland, OR", "Chicago, IL",
			"Los Angeles, CA", "Miami, FL", "Atlanta, GA", "Phoenix, AZ",
		},
		schools: []string{
			"MIT", "Stanford University", "UC Berkeley", "Carnegie Mellon",
			"Georgia Tech", "University of Washington", "Cornell University",
			"University of Texas", "Purdue University", "UCLA",
		},
		degrees: []string{
			"BS Computer Science", "MS Computer Science", "BS Software Engineering",
			"MS Data Science", "BS Information Technology", "PhD Computer Science",
		},
	}
}

// Generate creates a random persona.
func (g *PersonaGenerator) Generate() *Persona {
	firstName := g.pick(g.firstNames)
	lastName := g.pick(g.lastNames)
	email := fmt.Sprintf("%s.%s@%s",
		strings.ToLower(firstName),
		strings.ToLower(lastName),
		g.pick(g.domains),
	)

	title := g.pick(g.titles)

	return &Persona{
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		Phone:      g.generatePhone(),
		Address:    g.generateAddress(),
		City:       g.pick(g.cities),
		Country:    "USA",
		Title:      title,
		Summary:    g.generateSummary(title),
		Experience: g.generateExperience(),
		Education:  g.generateEducation(),
		Skills:     g.pickN(g.skills, 8),
	}
}

// FullName returns the full name.
func (p *Persona) FullName() string {
	return p.FirstName + " " + p.LastName
}

func (g *PersonaGenerator) pick(items []string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(items))))
	return items[n.Int64()]
}

func (g *PersonaGenerator) pickN(items []string, n int) []string {
	if n > len(items) {
		n = len(items)
	}
	// Shuffle and take first n
	shuffled := make([]string, len(items))
	copy(shuffled, items)
	for i := len(shuffled) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		shuffled[i], shuffled[j.Int64()] = shuffled[j.Int64()], shuffled[i]
	}
	return shuffled[:n]
}

func (g *PersonaGenerator) generatePhone() string {
	area, _ := rand.Int(rand.Reader, big.NewInt(800))
	area = area.Add(area, big.NewInt(200))
	pre, _ := rand.Int(rand.Reader, big.NewInt(900))
	pre = pre.Add(pre, big.NewInt(100))
	line, _ := rand.Int(rand.Reader, big.NewInt(9000))
	line = line.Add(line, big.NewInt(1000))
	return fmt.Sprintf("(%d) %d-%d", area, pre, line)
}

func (g *PersonaGenerator) generateAddress() string {
	num, _ := rand.Int(rand.Reader, big.NewInt(9900))
	num = num.Add(num, big.NewInt(100))
	streets := []string{"Oak St", "Maple Ave", "Tech Blvd", "Innovation Way", "Code Lane", "Data Dr"}
	return fmt.Sprintf("%d %s", num, g.pick(streets))
}

func (g *PersonaGenerator) generateSummary(title string) string {
	years, _ := rand.Int(rand.Reader, big.NewInt(10))
	years = years.Add(years, big.NewInt(3))
	return fmt.Sprintf("Experienced %s with %d+ years building scalable systems.", title, years)
}

func (g *PersonaGenerator) generateExperience() []Job {
	count, _ := rand.Int(rand.Reader, big.NewInt(2))
	count = count.Add(count, big.NewInt(2))

	jobs := make([]Job, count.Int64())
	for i := range jobs {
		jobs[i] = Job{
			Company:  g.pick(g.companies),
			Title:    g.pick(g.titles),
			Duration: g.generateDuration(),
			Desc:     "Developed and maintained critical systems.",
		}
	}
	return jobs
}

func (g *PersonaGenerator) generateDuration() string {
	starts := []string{"2020", "2019", "2018", "2017", "2016"}
	ends := []string{"Present", "2023", "2022", "2021"}
	return g.pick(starts) + " - " + g.pick(ends)
}

func (g *PersonaGenerator) generateEducation() []School {
	return []School{
		{
			Name:   g.pick(g.schools),
			Degree: g.pick(g.degrees),
			Year:   "2018",
		},
	}
}
