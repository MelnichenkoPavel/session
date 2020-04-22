**Точки доступа**
1. главная
`curl -X GET http://127.0.0.1:8088/`
2. чтение
`curl -X GET 'http://127.0.0.1:8088/read?key=1'`
3. запись
`curl -X POST http://localhost:8088/write -H 'content-type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW' -F key=1 -F 'value={"a":1}'`

**Запуск**
1. переходим в каталог с тестируемой БД (например cd ./mongodb)
2. запускаем контейнера (docker-compose up -d)
3. выполняем запросы

MONGODB
1) mongo --username root --password example --authenticationDatabase admin
2) use db1
3) db.collection.find({ key: "1" }).pretty() 


**Нагрузочное**

`docker run -v $(pwd)/loadtest:/var/loadtest --net host -it direvius/yandex-tank -c custom-config-name.yaml`