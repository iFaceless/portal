package portal

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ifaceless/portal/field"
)

type ManagerModel struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *ManagerModel) Fullname() string {
	return m.Name + "xixi_haha"
}

type CompanyModel struct {
	ID        int
	ManagerID int
	Name      string
	Addr      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *CompanyModel) Manager() *ManagerModel {
	// perform a db query, and return result
	time.Sleep(5 * time.Millisecond)
	return &ManagerModel{
		ID:        c.ManagerID,
		Name:      fmt.Sprintf("manager_%d", time.Now().Unix()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type ProductModel struct {
	ID        int
	CompanyID int
	Name      string
	Price     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *ProductModel) Company() *CompanyModel {
	// perform a db query, and return result
	time.Sleep(5 * time.Millisecond)
	return &CompanyModel{
		ID:        p.CompanyID,
		ManagerID: rand.Intn(1024) + 1,
		Name:      fmt.Sprintf("company_%d", rand.Intn(100)),
		Addr:      "addr_company",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func makeProducts(count int) (ret []*ProductModel) {
	for i := 0; i < count; i++ {
		ret = append(ret, &ProductModel{
			ID:        i,
			CompanyID: i + rand.Intn(100),
			Name:      fmt.Sprintf("name_%d", i),
			Price:     i * 100,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return
}

type ManagerSchema struct {
	ID        string           `json:"id"`
	Name      string           `json:"name" portal:"attr:Fullname"`
	CreatedAt *field.Timestamp `json:"created_at"`
	UpdatedAt *field.Timestamp `json:"updated_at"`
}

type CompanySchema struct {
	ID        string           `json:"id,omitempty"`
	Manager   *ManagerSchema   `json:"manager,omitempty"`
	Name      string           `json:"name,omitempty"`
	Addr      string           `json:"addr,omitempty" portal:"meth:GetAddr"`
	CreatedAt *field.Timestamp `json:"created_at,omitempty"`
	UpdatedAt *field.Timestamp `json:"updated_at,omitempty"`
}

func (c *CompanySchema) GetAddr(company *CompanyModel) string {
	return fmt.Sprintf("custom_%s", company.Addr)
}

type ProductSchema struct {
	ID        string           `json:"id,omitempty"`
	Company   *CompanySchema   `json:"company_id,omitempty"`
	Name      string           `json:"name,omitempty"`
	Price     int              `json:"price,omitempty"`
	CreatedAt *field.Timestamp `json:"created_at,omitempty"`
	UpdatedAt *field.Timestamp `json:"updated_at,omitempty"`
}

// BenchmarkDumpManyLargeWorkerPool-4   	     165	   7373380 ns/op
func BenchmarkDumpManyLargeWorkerPool(b *testing.B) {
	SetMaxPoolSize(1000)
	products := makeProducts(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas []*ProductSchema
		_ = Dump(&schemas, products)
	}
}

// BenchmarkDumpOneLargeWorkerPool-4   	     199	   6059241 ns/op
func BenchmarkDumpOneLargeWorkerPool(b *testing.B) {
	SetMaxPoolSize(1000)
	products := makeProducts(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas *ProductSchema
		_ = Dump(&schemas, products[0])
	}
}

// BenchmarkDumpManyIgnoreDBQueryLargeWorkerPool-4   	   11694	     99205 ns/op
func BenchmarkDumpManyIgnoreDBQueryLargeWorkerPool(b *testing.B) {
	SetMaxPoolSize(1000)
	products := makeProducts(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas []*ProductSchema
		_ = Dump(&schemas, products, Exclude("Company"))
	}
}

// BenchmarkDumpOneIgnoreDBQueryLargeWorkerPool-4   	   92692	     11246 ns/op
func BenchmarkDumpOneIgnoreDBQueryLargeWorkerPool(b *testing.B) {
	SetMaxPoolSize(1000)
	products := makeProducts(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas *ProductSchema
		_ = Dump(&schemas, products[0], Exclude("Company"))
	}
}

// BenchmarkDumpManyIgnoreDBQuerySmallWorkerPool-4   	    7856	    144661 ns/op
func BenchmarkDumpManyIgnoreDBQuerySmallWorkerPool(b *testing.B) {
	SetMaxPoolSize(1)
	products := makeProducts(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas []*ProductSchema
		_ = Dump(&schemas, products, Exclude("Company"))
	}
}

// BenchmarkDumpOneIgnoreDBQuerySmallWorkerPool-4   	  110576	     11076 ns/op
func BenchmarkDumpOneIgnoreDBQuerySmallWorkerPool(b *testing.B) {
	SetMaxPoolSize(1)
	products := makeProducts(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schemas *ProductSchema
		_ = Dump(&schemas, products[0], Exclude("Company"))
	}
}
