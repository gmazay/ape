# Ape
**Configurable REST API server**

Маршруты, запросы, принимаемые и возвращаемые параметры,
тип авторизации и прочее - задаются в конфиге в форматах YAML или JSON.

Если имя файла имеет суффикс .yml - он парсится как YAML, иначе - JSON.

Пример: etc/ape.yml

Дефолтный хендлер  обрабатывает параметры из формы и возвращает данные в 2 форматах:

JSON (по умолчанию) или CSV (&format=csv)

Если в запрос добавить параметр download, будет задан Context-Type: application/download


В комплекте ткаже имеется обработчик FS() - файловый сервер.


Если функционала базовых обработчиков недостаточно, нужно:
```
1) добавить свой в controller/custom.go
2) добавить ссылку на него в Dispatcher(), controller/base.go
3) указать handler в конфиге
```


**Get packages & building:**
```
go get github.com/valyala/fasthttp
go get github.com/buaazp/fasthttprouter
go get github.com/kpango/glg
go get github.com/lib/pq
go get gopkg.in/yaml.v2

go build -o ape main.go
```

**Start:**
```
./ape -conf ape.yml
```


**Configure systemd:**

/etc/systemd/system/ape.service
```ini
[Unit]
Description=Ape REST server
After=network.target
After=mariadb.service

[Service]
PIDFile=/var/run/ape/ape.pid
ExecStart=/opt/ape/ape -conf /opt/ape/ape.yml
ExecStop=/bin/kill -SIGTERM $MAINPID
ExecReload=/bin/kill -SIGURG $MAINPID
Restart=always
KillSignal=SIGQUIT
Type=simple
StandardError=syslog
NotifyAccess=all
WorkingDirectory=/opt/ape

User=ape
Group=ape

[Install]
WantedBy=multi-user.target
```
systemctl enable ape

systemctl start ape

 Soft reload config:

systemctl reload ape / service ape reload

