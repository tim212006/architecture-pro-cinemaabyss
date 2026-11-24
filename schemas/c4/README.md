# C4-диаграммы CinemaAbyss

Диаграммы выполнены в формате C4, как в проектной работе спринта 1.

## Уровень 1. Диаграмма контекста

- [[context-to-be.puml|Исходник PlantUML]]

![[context-to-be.png|Контекстная диаграмма C4 To-Be]]

Диаграмма показывает CinemaAbyss как единую систему и её окружение: пользователей, внешний Identity Provider, платёжный сервис, поставщиков контента и рекомендательную систему.

## Уровень 2. Диаграмма контейнеров

- [[container-to-be.puml|Исходник PlantUML]]

![[container-to-be.png|Диаграмма контейнеров C4 To-Be]]

Диаграмма показывает API Gateway, доменные микросервисы, Kafka, `events-service` и монолит как временный миграционный контейнер.

## Уровень 3. Диаграммы компонентов

### proxy-service

- [[component-proxy-service.puml|Исходник PlantUML]]

![[component-proxy-service.png|Компоненты proxy-service]]

### movies-service

- [[component-movies-service.puml|Исходник PlantUML]]

![[component-movies-service.png|Компоненты movies-service]]

### subscriptions-service

- [[component-subscriptions-service.puml|Исходник PlantUML]]

![[component-subscriptions-service.png|Компоненты subscriptions-service]]

### ratings-service

- [[component-ratings-service.puml|Исходник PlantUML]]

![[component-ratings-service.png|Компоненты ratings-service]]

### user-collections-service

- [[component-user-collections-service.puml|Исходник PlantUML]]

![[component-user-collections-service.png|Компоненты user-collections-service]]

### events-service

- [[component-events-service.puml|Исходник PlantUML]]

![[component-events-service.png|Компоненты events-service]]
