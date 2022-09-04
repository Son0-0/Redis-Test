# Redis-Test

<img width="799" alt="스크린샷 2022-09-05 오전 1 13 32" src="https://user-images.githubusercontent.com/81317358/188323047-94af1379-3044-44fe-9780-d31cc9bf28e5.png">

로컬 테스트:
```
[방법1]
git clone https://github.com/Son0-0/Redis-Test
cd Go
make docker

[방법2]
git clone https://github.com/Son0-0/Redis-Test
cd Go
docker compose up --build
```

Request command:

DB에서 데이터 접근 테스트

```
curl http://localhost:9090/db?codes=KRWUSD | jq
or
curl http://localhost:9090/db?codes=KRWUSD
```

Open API fetch 테스트

```
curl http://localhost:9090/api?codes=KRWUSD | jq
or
curl http://localhost:9090/api?codes=KRWUSD
```

<img width="1480" alt="스크린샷 2022-09-05 오전 3 54 34" src="https://user-images.githubusercontent.com/81317358/188329235-b44d045e-9819-46ad-9638-09e0a4382057.png">

## Caching
- 원본 데이터에 접근하는 시간보다 캐시 데이터에 접근하는 시간이 빨라야 함
- DB 접근 시간 or 또 다른 API 호출을 통한 데이터 접근 시간보다 짧아야 함
- 동일한 데이터에 대해 반복적으로 접근하는 상황이 많을 때 사용
  - 데이터의 재사용 횟수가 1회 이상
