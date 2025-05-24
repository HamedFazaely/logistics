## Logistics Service
This service is responsible to do CRUD operations for customers, orders and providers. It also notifies sms service to send sms for pickedup orders and updates order status through messages received from satts service

## How to run
Build the image:
`docker build -t logistics:v1.0.0 .`

Create a docker network: `docker network create -d bridge podro`
Run mysql docker container and create the database and table whose definitions are in db_tables.txt at the root of the project: `docker run --name podro-mysql --network=podro -v /path/to/data:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=rootpass -p 3306:3306 -d mysql:9.3.0`
Run rabbitmq docker container: `docker run -d --rm --name rabbitmq --network=podro -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=user -e RABBITMQ_DEFAULT_PASS=password rabbitmq:4-management`

Run logistics container: `docker run -d --rm --name logist --network=podro -p 8080:8080 -e DB_USER=dbuser -e DB_HOST=podro-mysql -e DB_PASSWORD=dbpass -e MQ_HOST=rabbitmq -e MQ_PASSWORD=mqpass -e MQ_USER=mquser podro/logistics:v1.0.0`
