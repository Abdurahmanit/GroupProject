package entity

import (
	"errors"
	"time"
)

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func NewCartItem(productID string, quantity int) (*CartItem, error) {
	if productID == "" {
		return nil, errors.New("product ID cannot be empty for cart item")
	}
	if quantity <= 0 {
		return nil, errors.New("cart item quantity must be positive")
	}
	return &CartItem{ProductID: productID, Quantity: quantity}, nil
}

type Cart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func NewCart(userID string) *Cart {
	return &Cart{
		UserID:    userID,
		Items:     make([]CartItem, 0),
		UpdatedAt: time.Now().UTC(),
	}
}

func (c *Cart) GetItem(productID string) (*CartItem, int) {
	for i, item := range c.Items {
		if item.ProductID == productID {
			return &c.Items[i], i
		}
	}
	return nil, -1
}

func (c *Cart) AddItem(productID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity to add must be positive")
	}

	item, _ := c.GetItem(productID)
	if item != nil {
		item.Quantity += quantity
	} else {
		newItem, err := NewCartItem(productID, quantity)
		if err != nil {
			return err
		}
		c.Items = append(c.Items, *newItem)
	}
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Cart) UpdateItemQuantity(productID string, newQuantity int) error {
	item, index := c.GetItem(productID)
	if item == nil {
		return errors.New("item not found in cart")
	}

	if newQuantity <= 0 {
		c.Items = append(c.Items[:index], c.Items[index+1:]...)
	} else {
		item.Quantity = newQuantity
	}
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Cart) RemoveItem(productID string) error {
	_, index := c.GetItem(productID)
	if index == -1 {
		return errors.New("item not found in cart to remove")
	}

	c.Items = append(c.Items[:index], c.Items[index+1:]...)
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Cart) Clear() {
	c.Items = make([]CartItem, 0)
	c.UpdatedAt = time.Now().UTC()
}
