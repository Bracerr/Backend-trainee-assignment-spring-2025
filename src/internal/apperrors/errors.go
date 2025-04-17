package apperrors

import "errors"

var (
	ErrUserAlreadyExists      = errors.New("пользователь уже существует")
	ErrInvalidCredentials     = errors.New("неверные учетные данные")
	ErrInvalidRole            = errors.New("недопустимая роль")
	ErrValidationFailed       = errors.New("неверные данные")
	ErrInvalidCity            = errors.New("недопустимый город")
	ErrActiveReceptionExists  = errors.New("уже есть активная приемка")
	ErrNoActiveReception      = errors.New("нет активной приемки")
	ErrInvalidProductType     = errors.New("недопустимый тип товара")
	ErrPVZNotFound            = errors.New("ПВЗ не найден")
	ErrReceptionClosed        = errors.New("приемка закрыта")
	ErrProductNotLast         = errors.New("можно удалить только последний добавленный товар")
	ErrNoProductsToDelete     = errors.New("нет товаров для удаления")
	ErrNoProductsInReception  = errors.New("нет товаров в приемке")
	ErrReceptionAlreadyClosed = errors.New("приемка уже закрыта")
)
