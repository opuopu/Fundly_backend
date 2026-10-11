package category

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryName = errors.New("category name cannot be empty")
	ErrInvalidCategoryType = errors.New("invalid category type: must be income or expense")
	ErrNoFieldsToUpdate    = errors.New("at least one field must be provided to update")
)

type CategoryType string

const (
	TypeIncome  CategoryType = "INCOME"
	TypeExpense CategoryType = "EXPENSE"
)

func ParseCategoryType(s string) (CategoryType, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "INCOME":
		return TypeIncome, nil
	case "EXPENSE":
		return TypeExpense, nil
	default:
		return "", ErrInvalidCategoryType
	}
}

type Category struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      CategoryType `json:"type"`
	Icon      string       `json:"icon"`
	Color     string       `json:"color"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewCategory(id, name, catType, icon, color string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidCategoryName
	}

	t, err := ParseCategoryType(catType)
	if err != nil {
		return nil, err
	}

	icon = strings.TrimSpace(icon)
	if icon == "" {
		icon = "tag"
	}

	color = strings.TrimSpace(color)
	if color == "" {
		color = "#2196F3"
	}

	return &Category{
		ID:        id,
		Name:      name,
		Type:      t,
		Icon:      icon,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (c *Category) Update(name, color, icon *string) error {
	if name == nil && color == nil && icon == nil {
		return ErrNoFieldsToUpdate
	}
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return ErrInvalidCategoryName
		}
		c.Name = trimmedName
	}
	if color != nil {
		trimmedColor := strings.TrimSpace(*color)
		if trimmedColor != "" {
			c.Color = trimmedColor
		}
	}
	if icon != nil {
		trimmedIcon := strings.TrimSpace(*icon)
		if trimmedIcon != "" {
			c.Icon = trimmedIcon
		}
	}
	return nil
}
