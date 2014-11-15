package main

import (
	"time"
)

type Person struct {
	Id         string     `json:"id"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
	ExternalId string     `json:"externalId"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
}

type Tenant struct {
	Id        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatorId string     `json:"creatorId"`
	Name      string     `json:"name"`
	Title     string     `json:"title"`
	Domain    string     `json:"domain"`

	Creator     *Person      `json:"creator,omitempty" sql:"-"`
	Memberships []Membership `json:"memberships"`
}

type Client struct {
	Id        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatorId string     `json:"creatorId"`
	TenantId  string     `json:"tenantId"`
	Name      string     `json:"name"`

	Creator     *Person      `json:"creator,omitempty" sql:"-"`
	Tenant      *Tenant      `json:"tenant,omitempty" sql:"-"`
	Projects    []Project    `json:"projects"`
	Memberships []Membership `json:"memberships"`
}

type Project struct {
	Id        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatorId string     `json:"creatorId"`
	ClientId  string     `json:"clientId"`
	Name      string     `json:"name"`

	Creator     *Person      `json:"creator,omitempty" sql:"-"`
	Client      *Client      `json:"client,omitempty" sql:"-"`
	Memberships []Membership `json:"memberships"`
}

type Review struct {
	Id        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatorId string     `json:"creatorId"`
	ProjectId string     `json:"projectId"`
	Name      string     `json:"name"`
	Message   string     `json:"message"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expiresAt"`

	Creator     *Person      `json:"creator,omitempty" sql:"-"`
	Project     *Project     `json:"project,omitempty" sql:"-"`
	Memberships []Membership `json:"memberships"`
}

type Membership struct {
	Id        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatorId string     `json:"creatorId"`
	PersonId  *string    `json:"personId"`
	Type      string     `json:"type"`
	TenantId  *string    `json:"tenantId"`
	ClientId  *string    `json:"clientId"`
	ProjectId *string    `json:"projectId"`
	ReviewId  *string    `json:"reviewId"`

	Creator *Person  `json:"creator,omitempty" sql:"-"`
	Person  *Person  `json:"person,omitempty" sql:"-"`
	Tenant  *Tenant  `json:"tenant,omitempty" sql:"-"`
	Client  *Client  `json:"client,omitempty" sql:"-"`
	Project *Project `json:"project,omitempty" sql:"-"`
	Review  *Review  `json:"review,omitempty" sql:"-"`
}
