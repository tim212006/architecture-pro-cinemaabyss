## Изучите [README.md](README.md) файл и структуру проекта.

## Задание 1

1. Спроектируйте to be архитектуру КиноБездны, разделив всю систему на отдельные домены и организовав интеграционное взаимодействие и единую точку вызова сервисов.
Результат представьте в виде контейнерной диаграммы в нотации С4.
Добавьте ссылку на файл в этот шаблон
[[ссылка|ссылка на файл]]

**Выполнено:**
Каталог диаграмм: [[Спринт 2 Разбивка монолитной системы на микросервисы, работа с инфраструктурой/Проектная работа 2/architecture-pro-cinemaabyss/schemas/c4/README|schemas/c4]]

### Уровень 1 — контекст

- [[Спринт 2 Разбивка монолитной системы на микросервисы, работа с инфраструктурой/Проектная работа 2/architecture-pro-cinemaabyss/schemas/c4/context-to-be.puml|Исходник PlantUML]]

![[Спринт 2 Разбивка монолитной системы на микросервисы, работа с инфраструктурой/Проектная работа 2/architecture-pro-cinemaabyss/schemas/c4/context-to-be.png|Контекстная диаграмма C4 To-Be]]

### Уровень 2 — контейнеры

- [[Спринт 2 Разбивка монолитной системы на микросервисы, работа с инфраструктурой/Проектная работа 2/architecture-pro-cinemaabyss/schemas/c4/container-to-be.puml|Исходник PlantUML]]

![[Спринт 2 Разбивка монолитной системы на микросервисы, работа с инфраструктурой/Проектная работа 2/architecture-pro-cinemaabyss/schemas/c4/container-to-be.png|Диаграмма контейнеров C4 To-Be]]

### Уровень 3 — компоненты

- [[component-proxy-service.puml|proxy-service]]
- [[component-movies-service.puml|movies-service]]
- [[component-subscriptions-service.puml|subscriptions-service]]
- [[component-ratings-service.puml|ratings-service]]
- [[component-user-collections-service.puml|user-collections-service]]
- [[component-events-service.puml|events-service]]

## Задание 2

### 1. Proxy
Команда КиноБездны уже выделила сервис метаданных о фильмах movies и вам необходимо реализовать бесшовный переход с применением паттерна Strangler Fig в части реализации прокси-сервиса (API Gateway), с помощью которого можно будет постепенно переключать траффик, используя фиче-флаг.


Реализуйте сервис на любом языке программирования в ./src/microservices/proxy.
Конфигурация для запуска сервиса через docker-compose уже добавлена
```yaml
  proxy-service:
    build:
      context: ./src/microservices/proxy
      dockerfile: Dockerfile
    container_name: cinemaabyss-proxy-service
    depends_on:
      - monolith
      - movies-service
      - events-service
    ports:
      - "8000:8000"
    environment:
      PORT: 8000
      MONOLITH_URL: http://monolith:8080
      #монолит
      MOVIES_SERVICE_URL: http://movies-service:8081 #сервис movies
      EVENTS_SERVICE_URL: http://events-service:8082 
      GRADUAL_MIGRATION: "true" # вкл/выкл простого фиче-флага
      MOVIES_MIGRATION_PERCENT: "50" # процент миграции
    networks:
      - cinemaabyss-network
```

- После реализации запустите postman тесты - они все должны быть зеленые.
- Отправьте запросы к API Gateway:
   ```bash
   curl http://localhost:8000/api/movies
   ```
- Протестируйте постепенный переход, изменив переменную окружения MOVIES_MIGRATION_PERCENT в файле docker-compose.yml.

#### **Выполнено:**
Proxy реализует паттерн Strangler Fig и выступает единой точкой входа для клиентских запросов.

Маршрутизация выполняется следующим образом:

- `GET /health` обрабатывается самим proxy;
- запросы к `/api/users` направляются в монолит;
- запросы к `/api/movies` направляются в монолит или сервис `movies`;
- при `GRADUAL_MIGRATION=false` весь трафик `/api/movies` направляется в монолит;
- при `GRADUAL_MIGRATION=true` маршрут выбирается с учётом `MOVIES_MIGRATION_PERCENT`.

Конфигурация сервисов передаётся через переменные окружения:

- `MONOLITH_URL`;
- `MOVIES_SERVICE_URL`;
- `GRADUAL_MIGRATION`;
- `MOVIES_MIGRATION_PERCENT`.

Проверка выполнена локально:

- при значении миграции `0%` запросы `/api/movies` направлялись в монолит;
- при значении миграции `100%` запросы `/api/movies` направлялись в `movies`;
- при значении миграции `50%` использовались оба маршрута;
- при отключённом `GRADUAL_MIGRATION` весь трафик направлялся в монолит;
- запросы через proxy возвращали успешный ответ и корректно сохраняли query-параметр `id`.

Команда запуска для проверки:
```bash
sudo docker run --rm \
  --name cinemaabyss-proxy-test \
  -p 8000:8000 \
  -e PORT=8000 \
  -e MONOLITH_URL=http://host.docker.internal:8080 \
  -e MOVIES_SERVICE_URL=http://host.docker.internal:8081 \
  -e GRADUAL_MIGRATION=false \
  -e MOVIES_MIGRATION_PERCENT=0 \
  --add-host=host.docker.internal:host-gateway \
  cinemaabyss-proxy-test
```


Команда проверки API Gateway:

```bash
curl http://localhost:8000/api/movies
```

### 2. Kafka
 Вам как архитектуру нужно также проверить гипотезу насколько просто реализовать применение Kafka в данной архитектуре.

Для этого нужно сделать MVP сервис events, который будет при вызове API создавать и сам же читать сообщения в топике Kafka.

    - Разработайте сервис на любом языке программирования с consumer'ами и producer'ами.
    - Реализуйте простой API, при вызове которого будут создаваться события User/Payment/Movie и обрабатываться внутри сервиса с записью в лог
    - Добавьте в docker-compose новый сервис, kafka там уже есть

Необходимые тесты для проверки этого API вызываются при запуске npm run test:local из папки tests/postman 
Приложите скриншот тестов и скриншот состояния топиков Kafka http://localhost:8090 


#### **Выполнено:**

Реализован MVP-сервис `events` на Go.

Сервис поддерживает следующие endpoint’ы:

- `GET /api/events/health`;
- `POST /api/events/movie`;
- `POST /api/events/user`;
- `POST /api/events/payment`.

Producer формирует событие с полями:

- `id`;
- `type`;
- `timestamp`;
- `payload`.

События публикуются в соответствующие Kafka topic’и:

- `movie-events`;
- `user-events`;
- `payment-events`.

В сервисе запущены consumer’ы для всех трёх topic’ов. После получения сообщения consumer декодирует событие и записывает информацию о нём в лог:

```text
Event processed: topic=movie-events partition=0 offset=2 ...
````

Проверка выполнена локально:

- события фильма, пользователя и платежа возвращают статус `201 Created`;
- ответ содержит `status`, `partition`, `offset` и `event`;
- consumer’ы запускаются для всех трёх topic’ов;
- обработанные события появляются в логах `events-service`;
- локальные Postman-тесты завершились успешно: 22 запроса, 42 проверки, 0 ошибок.

Скриншот успешного запуска Postman-тестов:

[[Tests.png]]
	![[Tests.png|Postman tests]]

Скриншот состояния Kafka topic’ов:

[[Topics.png]]
![[Topics.png|KafkaUI]]

## Задание 3

Команда начала переезд в Kubernetes для лучшего масштабирования и повышения надежности. 
Вам, как архитектору осталось самое сложное:
 - реализовать CI/CD для сборки прокси сервиса
 - реализовать необходимые конфигурационные файлы для переключения трафика.


### CI/CD

 В папке .github/worflows доработайте деплой новых сервисов proxy и events в docker-build-push.yml , чтобы api-tests при сборке отрабатывали корректно при отправке коммита в вашу новую ветку.

Нужно доработать 
```yaml
on:
  push:
    branches: [ main ]
    paths:
      - 'src/**'
      - '.github/workflows/docker-build-push.yml'
  release:
    types: [published]
```
и добавить необходимые шаги в блок
```yaml
jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to the Container registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

```
Как только сборка отработает и в github registry появятся ваши образы, можно переходить к блоку настройки Kubernetes
Успешным результатом данного шага является "зеленая" сборка и "зеленые" тесты

**Выполнено:**

[[BuildAndPush.png]]

![[BuildAndPush.png|Результат сборки]]



### Proxy в Kubernetes

#### Шаг 1
Для деплоя в kubernetes необходимо залогиниться в docker registry Github'а.
1. Создайте Personal Access Token (PAT) https://github.com/settings/tokens . Создавайте class с правом read:packages

**Выполнено:**
Создан классический токен с правом read:packges

2. В src/kubernetes/*.yaml (event-service, monolith, movies-service и proxy-service)  отредактируйте путь до ваших образов 
```bash
 spec:
      containers:
      - name: events-service
        image: ghcr.io/ваш логин/имя репозитория/events-service:latest
```

**Выполнено:**
Отредактирован путь до моих образов:
```bash
grep -RIn --include='*.yaml' 'image:' \
  src/kubernetes/events-service.yaml \
  src/kubernetes/monolith.yaml \
  src/kubernetes/movies-service.yaml \
  src/kubernetes/proxy-service.yaml

src/kubernetes/events-service.yaml:22:          image: ghcr.io/tim212006/architecture-pro-cinemaabyss/events-service:latest
src/kubernetes/monolith.yaml:25:        image:  ghcr.io/tim212006/architecture-pro-cinemaabyss/monolith:latest
src/kubernetes/movies-service.yaml:25:        image: ghcr.io/tim212006/architecture-pro-cinemaabyss/monolith:latest
src/kubernetes/proxy-service.yaml:22:          image: ghcr.io/tim212006/architecture-pro-cinemaabyss/proxy-service:latest

```

3. Добавьте в секрет src/kubernetes/dockerconfigsecret.yaml в поле
```bash
 .dockerconfigjson: значение в base64 файла ~/.docker/config.json
```


**Выполнено.** Secret `dockerconfigjson` сформирован из `~/.docker/config.json`, проверен через Kubernetes API и создан в namespace `cinemaabyss`.

Проверка:

```text
secret/dockerconfigjson created
```

4. Если в ~/.docker/config.json нет значения для аутентификации
```json
{
        "auths": {
                "ghcr.io": {
                       тут пусто
                }
        }
}
```
то выполните 

и добавьте

```json 
 "auth": "имя пользователя:токен в base64"
```

Чтобы получить значение в base64 можно выполнить команду
```bash
 echo -n ваш_логин:ваш_токен | base64
```

После заполнения config.json, также прогоните содержимое через base64

```bash
cat .docker/config.json | base64
```

и полученное значение добавляем в

```bash
 .dockerconfigjson: значение в base64 файла ~/.docker/config.json
```

#### Шаг 2

  Доработайте src/kubernetes/event-service.yaml и src/kubernetes/proxy-service.yaml

  - Необходимо создать Deployment и Service 
  - Доработайте ingress.yaml, чтобы можно было с помощью тестов проверить создание событий
  - Выполните дальшейшие шаги для поднятия кластера:

  1. Создайте namespace:
  ```bash
  kubectl apply -f src/kubernetes/namespace.yaml
  ```

**Выполнено:** 
Namespace `cinemaabyss` создан и активен.

```text
NAME          STATUS   AGE
cinemaabyss   Active   80m
```

2. Создайте секреты и переменные
  ```bash
  kubectl apply -f src/kubernetes/configmap.yaml
  kubectl apply -f src/kubernetes/secret.yaml
  kubectl apply -f src/kubernetes/dockerconfigsecret.yaml
  kubectl apply -f src/kubernetes/postgres-init-configmap.yaml
  ```

**Выполнено:** В namespace `cinemaabyss` созданы ConfigMap и Secret:

```bash
NAME                              DATA   AGE
configmap/cinemaabyss-config      8      68m
configmap/kube-root-ca.crt        1      86m
configmap/postgres-init-scripts   1      23s

NAME                         TYPE                             DATA   AGE
secret/cinemaabyss-secrets   Opaque                           1      24s
secret/dockerconfigjson      kubernetes.io/dockerconfigjson   1      24m

```

  3. Разверните базу данных:
  ```bash
  kubectl apply -f src/kubernetes/postgres.yaml
  ```

  На этом этапе если вызвать команду
  ```bash
  kubectl -n cinemaabyss get pod
  ```
  Вы увидите

  NAME         READY   STATUS    
  postgres-0   1/1     Running   
  
**Выполнено:** PostgreSQL развернут в namespace `cinemaabyss`.

```bash
postgres-0                       1/1     Running   0          4m46s
```


  4. Разверните Kafka:
  ```bash
  kubectl apply -f src/kubernetes/kafka/kafka.yaml
  ```

  Проверьте, теперь должно быть запущено 3 пода, если что-то не так, то посмотрите логи
  ```bash
  kubectl -n cinemaabyss logs имя_пода (например - kafka-0)
  ```

**Выполнено:**

```bash
NAME                             READY   STATUS    RESTARTS   AGE
kafka-0                          1/1     Running   0          25m
```

  5. Разверните монолит:
  ```bash
  kubectl apply -f src/kubernetes/monolith.yaml
  ```

**Выполнено:**

```bash
NAME                             READY   STATUS    RESTARTS   AGE
monolith-64b8559bc7-pdzgk        1/1     Running   0          22s
```
  
  6. Разверните микросервисы:
  ```bash
  kubectl apply -f src/kubernetes/movies-service.yaml
  kubectl apply -f src/kubernetes/events-service.yaml
  ```

**Выполнено:**
```bash
NAME                              READY   STATUS    RESTARTS   AGE
events-service-6d888f95cc-9vqhb   1/1     Running   0          10m
movies-service-845b76cdf8-f9xmn   1/1     Running   0          45s
```
  
  7. Разверните прокси-сервис:
  ```bash
  kubectl apply -f src/kubernetes/proxy-service.yaml
  ```

  После запуска и поднятия подов вывод команды 
  ```bash
  kubectl -n cinemaabyss get pod
  ```

  Будет наподобие такого

  NAME                              READY   STATUS    

  events-service-7587c6dfd5-6whzx   1/1     Running  

  kafka-0                           1/1     Running   

  monolith-8476598495-wmtmw         1/1     Running  

  movies-service-6d5697c584-4qfqs   1/1     Running  

  postgres-0                        1/1     Running  

  proxy-service-577d6c549b-6qfcv    1/1     Running  

  zookeeper-0                       1/1     Running 

**Выполнено:**
```bash
NAME                              READY   STATUS    RESTARTS   AGE
events-service-6d888f95cc-9vqhb   1/1     Running   0          13m
kafka-0                           1/1     Running   0          48m
monolith-64b8559bc7-pdzgk         1/1     Running   0          15m
movies-service-845b76cdf8-f9xmn   1/1     Running   0          3m57s
postgres-0                        1/1     Running   0          51m
proxy-service-84b8979fb-t8gqn     1/1     Running   0          31s
zookeeper-0                       1/1     Running   0          48m

```

  8. Добавим ingress

  - добавьте аддон
  ```bash
  minikube addons enable ingress
  ```
  ```bash
  kubectl apply -f src/kubernetes/ingress.yaml
  ```
  
**Выполнено:**
```bash
NAME                                       READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-7k5vx       0/1     Completed   0          52s
ingress-nginx-admission-patch-nspcb        0/1     Completed   0          52s
ingress-nginx-controller-d7cd8c989-wch28   1/1     Running     0          52s
```

```bash
Name:             cinemaabyss-ingress
Labels:           <none>
Namespace:        cinemaabyss
Address:          192.168.49.2
Ingress Class:    nginx
Default backend:  <default>
Rules:
  Host                     Path  Backends
  ----                     ----  --------
  cinemaabyss.example.com  
                           /             proxy-service:80 (10.244.0.16:8000)
                           /api/events   events-service:8082 (10.244.0.14:8082)
Annotations:               nginx.ingress.kubernetes.io/ssl-redirect: false
Events:
  Type    Reason  Age                  From                      Message
  ----    ------  ----                 ----                      -------
  Normal  Sync    28s (x3 over 4m35s)  nginx-ingress-controller  Scheduled for sync

```

  9. Добавьте в /etc/hosts
  127.0.0.1 cinemaabyss.example.com
  
  **Выполнено:**
 
  ```bash
  127.0.0.1       cinemaabyss.example.com
  ```

  10. Вызовите
  ```bash
  minikube tunnel
  ```
  
**Выполнено:**
```bash
minikube tunnel
Status:	
	machine: minikube
	pid: 259239
	route: 10.96.0.0/12 -> 192.168.49.2
	minikube: Running
	services: []
    errors: 
		minikube: no errors
		router: no errors
		loadbalancer emulator: no errors
```  
  
  11. Вызовите https://cinemaabyss.example.com/api/movies
  Вы должны увидеть вывод списка фильмов
  Можно поэкспериментировать со значением   MOVIES_MIGRATION_PERCENT в src/kubernetes/configmap.yaml и убедится, что вызовы movies уходят полностью в новый сервис

**Выполнено:**


![[ListOfMovies.png|Список фильмов]]

  12. Запустите тесты из папки tests/postman
  ```bash
   npm run test:kubernetes
  ```
  Часть тестов с health-чек упадет, но создание событий отработает.
  Откройте логи event-service и сделайте скриншот обработки событий

**Выполнено:**
![[LogsEventService.png|Логи тестового создания фильмов]]

#### Шаг 3
Добавьте сюда скриншота вывода при вызове https://cinemaabyss.example.com/api/movies и  скриншот вывода event-service после вызова тестов.

**Выполнено:**
https я не настраивал. Если нужно - сделаю 443 порт и настрою сертификаты.
Доступ по http

В п. 11 Скриншот общего списка.
В результате выполнения тестов создано две записи:

![[AddedMoviesAfterTests.png]]


## Задание 4
Для простоты дальнейшего обновления и развертывания вам как архитектуру необходимо так же реализовать helm-чарты для прокси-сервиса и проверить работу 

Для этого:
1. Перейдите в директорию helm и отредактируйте файл values.yaml

```yaml
# Proxy service configuration
proxyService:
  enabled: true
  image:
    repository: ghcr.io/db-exp/cinemaabysstest/proxy-service
    tag: latest
    pullPolicy: Always
  replicas: 1
  resources:
    limits:
      cpu: 300m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
  service:
    port: 80
    targetPort: 8000
    type: ClusterIP
```

- Вместо ghcr.io/db-exp/cinemaabysstest/proxy-service напишите свой путь до образа для всех сервисов
- для imagePullSecret проставьте свое значение (скопируйте из конфигурации kubernetes)
  ```yaml
  imagePullSecrets:
      dockerconfigjson: BASE64
  ```

2. В папке ./templates/services заполните шаблоны для proxy-service.yaml и events-service.yaml (опирайтесь на свою kubernetes конфигурацию - смысл helm'а сделать шаблоны для быстрого обновления и установки)

```yaml
template:
    metadata:
      labels:
        app: proxy-service
    spec:
      containers:
       Тут ваша конфигурация
```

3. Проверьте установку
Сначала удалим установку руками

```bash
kubectl delete all --all -n cinemaabyss
kubectl delete  namespace cinemaabyss
```
Запустите 
```bash
helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace
```
Если в процессе будет ошибка
```code
[2025-04-08 21:43:38,780] ERROR Fatal error during KafkaServer startup. Prepare to shutdown (kafka.server.KafkaServer)
kafka.common.InconsistentClusterIdException: The Cluster ID OkOjGPrdRimp8nkFohYkCw doesn't match stored clusterId Some(sbkcoiSiQV2h_mQpwy05zQ) in meta.properties. The broker is trying to join the wrong cluster. Configured zookeeper.connect may be wrong.
```

Проверьте развертывание:
```bash
kubectl get pods -n cinemaabyss
minikube tunnel
```

Потом вызовите 
https://cinemaabyss.example.com/api/movies
и приложите скриншот развертывания helm и вывода https://cinemaabyss.example.com/api/movies


**Выполнено:**
```bash
helm list -n cinemaabyss
kubectl get pods -n cinemaabyss
curl -i \
  -H 'Host: cinemaabyss.example.com' \
  http://$(minikube ip):30219/api/movies
NAME       	NAMESPACE  	REVISION	UPDATED                               	STATUS  	CHART            	APP VERSION
cinemaabyss	cinemaabyss	1       	2026-11-22 17:48:56.62422132 +0300 MSK	deployed	cinemaabyss-0.1.0	1.0.0      
NAME                              READY   STATUS    RESTARTS      AGE
events-service-6d888f95cc-zqsw6   1/1     Running   0             12m
kafka-0                           1/1     Running   0             12m
monolith-64b8559bc7-m55rb         1/1     Running   2 (12m ago)   12m
movies-service-845b76cdf8-csnkq   1/1     Running   2 (12m ago)   12m
postgres-0                        1/1     Running   0             12m
proxy-service-5fc54dbfbb-8s87w    1/1     Running   0             12m
zookeeper-0                       1/1     Running   0             12m
HTTP/1.1 200 OK
Date: Sun, 22 Nov 2026 15:01:41 GMT
Content-Type: application/json
Content-Length: 1267
Connection: keep-alive

[{"id":1,"title":"The Shawshank Redemption","description":"Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.","genres":["Drama"],"rating":9.3},{"id":2,"title":"The Godfather","description":"The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.","genres":["Crime","Drama"],"rating":9.2},{"id":3,"title":"The Dark Knight","description":"When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.","genres":["Action","Crime","Drama"],"rating":9},{"id":4,"title":"Pulp Fiction","description":"The lives of two mob hitmen, a boxer, a gangster and his wife, and a pair of diner bandits intertwine in four tales of violence and redemption.","genres":["Crime","Drama"],"rating":8.9},{"id":5,"title":"Forrest Gump","description":"The presidencies of Kennedy and Johnson, the Vietnam War, the Watergate scandal and other historical events unfold from the perspective of an Alabama man with an IQ of 75, whose only desire is to be reunited with his childhood sweetheart.","genres":["Drama","Romance"],"rating":8.8}]

```
# Задание 5
Компания планирует активно развиваться и для повышения надежности, безопасности, реализации сетевых паттернов типа Circuit Breaker и канареечного деплоя вам как архитектору необходимо развернуть istio и настроить circuit breaker для monolith и movies сервисов.

```bash

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

helm install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm install istio-ingressgateway istio/gateway -n istio-system
helm install istiod istio/istiod -n istio-system --wait

helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace

kubectl label namespace cinemaabyss istio-injection=enabled --overwrite

kubectl get namespace -L istio-injection

kubectl apply -f .\src\kubernetes\circuit-breaker-config.yaml -n cinemaabyss

```

Тестирование

# fortio
```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.25/samples/httpbin/sample-client/fortio-deploy.yaml -n cinemaabyss
```

# Get the fortio pod name
```bash
FORTIO_POD=$(kubectl get pod -n cinemaabyss | grep fortio | awk '{print $1}')

kubectl exec -n cinemaabyss $FORTIO_POD -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```
Например,

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg  -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```

Вывод будет типа такого

```bash
IP addresses distribution:
10.106.113.46:8081: 421
Code 200 : 79 (15.8 %)
Code 500 : 22 (4.4 %)
Code 503 : 399 (79.8 %)
```
Можно еще проверить статистику

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg -c istio-proxy -- pilot-agent request GET stats | grep movies-service | grep pending
```

И там смотрим 

```bash
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_total: 311 - столько раз срабатывал circuit breaker
You can see 21 for the upstream_rq_pending_overflow value which means 21 calls so far have been flagged for circuit breaking.
```

Приложите скриншот работы circuit breaker'а


Удаляем все
```bash
istioctl uninstall --purge
kubectl delete namespace istio-system
kubectl delete all --all -n cinemaabyss
kubectl delete namespace cinemaabyss
```


**Выполнено:**

![[CircuitBreaker.png]]

```bash
kubectl exec -n cinemaabyss \
  fortio-deploy-778485f875-jhmmd \
  -c istio-proxy -- \
  pilot-agent request GET stats \
  | grep 'movies-service' \
  | grep -E 'upstream_rq_(completed|200|503|pending_total|pending_overflow)|ejections_(detected_consecutive_5xx|enforced_consecutive_5xx|enforced_total)'
2026/11/22 16:05:22 INFO GOMEMLIMIT is already set, skipping package=github.com/KimMachineGun/automemlimit/memlimit GOMEMLIMIT=1073741824
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.external.upstream_rq_200: 7
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.external.upstream_rq_503: 493
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.external.upstream_rq_completed: 500
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.outlier_detection.ejections_detected_consecutive_5xx: 1
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.outlier_detection.ejections_enforced_consecutive_5xx: 1
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.outlier_detection.ejections_enforced_total: 1
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_200: 7
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_503: 493
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_completed: 500
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_overflow: 84
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_total: 8

```