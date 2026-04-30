# beforeMid — 최종발표 전 수정 목록

checklist.md 기준 전체 코드 점검 결과.

---

## Critical (먼저 고쳐야 함)

### 1. file.go — 파일 생성 후 DB INSERT 실패 시 orphan 파일 발생

`UploadFiles`에서 `os.Create()` 성공 후 DB INSERT 실패하면 디스크엔 파일이 남고 DB엔 없는 상태가 됨.

```
file.go:174  os.Create(...)                          → 디스크에 파일 생성
file.go:190  db.DB.Exec("INSERT INTO files ...")     → 실패 가능
```

**수정 방향**: DB INSERT 먼저 트랜잭션으로 실행, 전부 성공 후 파일 저장. 또는 INSERT 실패 시 `os.Remove()`로 파일 즉시 삭제.

---

### 2. share.go — 트랜잭션 내에서 DB.QueryRow() 직접 호출

`CreateShareLink` 트랜잭션 안에서 `db.DB.QueryRow()` 사용. tx 밖 커넥션으로 읽으면 트랜잭션 격리 깨짐.

```
share.go:131  db.DB.QueryRow("SELECT Permission ...")  ← tx 아님
```

**수정 방향**: `tx.QueryRow()`로 변경.

---

### 3. share.go — tx.Commit() 에러 미처리

```
share.go:149  tx.Commit()  ← 반환값 무시
```

**수정 방향**: `if err := tx.Commit(); err != nil { ... }` 추가.

---

### 4. user.go — JSON 디코딩 에러 무시

`Login` 핸들러에서 `Decode()` 에러를 확인하지 않음. 잘못된 body가 들어와도 그냥 진행.

```
user.go:84  json.NewDecoder(r.Body).Decode(&req)  ← 에러 미처리
```

Register(user.go:27)는 에러 체크함 — 불일치.

**수정 방향**: `if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(...) }` 통일.

---

### 5. common.go — generateID() 실패 시 log.Fatalf()로 서버 종료

`crypto/rand.Read()` 실패는 극히 드물지만, 발생하면 서버 프로세스 전체가 죽음. ID 생성 함수가 매 업로드/공유마다 호출되는 걸 감안하면 위험.

```
common.go:37  log.Fatalf("generateID: %v", err)
common.go:48  log.Fatalf("generateShareToken: %v", err)
```

**수정 방향**: 에러 반환으로 변경하고 호출부에서 처리.

---

## High (중요한 문제)

### 6. file.go — MIME 타입 실제 검증 없음

확장자 필터링(`blockedExts`)은 있지만, 파일 헤더의 `Content-Type`은 클라이언트가 임의로 보낼 수 있음.

```
file.go:119-121  blockedExts 체크 (확장자 기준)
file.go:128      fileHeader.Header.Get("Content-Type")  ← 클라이언트 제공, 신뢰 불가
```

**수정 방향**: 파일 첫 512바이트 읽어서 `http.DetectContentType()`으로 실제 타입 검증.

---

### 7. file.go — filepath.Clean() 미사용

`StoredName`은 UUID 기반이라 현재는 안전하지만, best practice 미준수.

```
file.go:172, 261, 306, 438, 472  filepath.Join(uploadFilesPath, ...)
```

**수정 방향**: `filepath.Clean(filepath.Join(uploadFilesPath, storedName))`으로 감싸고, 결과가 `uploadFilesPath` 하위인지 `strings.HasPrefix`로 검증.

---

### 8. file.go — rows.Close() 이중 호출

```
file.go:52   defer rows.Close()
file.go:64   rows.Close()         ← 이중 호출
```

defer가 이미 처리함. 명시적 호출 제거.

---

### 9. share.go — Rollback 누락 경로 있음

`tx.Exec()` 실패 후 return하는 경로에서 `tx.Rollback()` 없는 경우 있음.

```
share.go:283-285  return 전 Rollback 미호출
```

---

### 10. main.go — 포트 하드코딩

```
main.go:48  http.ListenAndServe(":8080", nil)
```

**수정 방향**: `const port = ":8080"` 상수화 또는 `os.Getenv("PORT")`.

---

## Medium (발표 전 여유되면)

### 11. file.go — UploadFiles 권한 INSERT 원자성

파일 INSERT 후 `file_permissions` INSERT 중 하나라도 실패하면 파일은 있는데 권한 없는 상태.

```
file.go:190-220  INSERT files → INSERT file_permissions (여러 건)
```

**수정 방향**: 전체를 하나의 트랜잭션으로 묶기.

---

### 12. common.go — AdminExists() DB 로직이 handler 패키지에 있음

checklist 기준: "핸들러는 요청/응답만, DB 로직은 db/ 패키지로".

```
common.go:58-72  AdminExists() 함수
```

**수정 방향**: `db.AdminExists()` 로 이동.

---

### 13. init.go — TOCTOU race condition

관리자 존재 여부 확인 → 생성 사이에 두 요청이 동시에 들어오면 관리자가 두 명 생성될 수 있음.

```
init.go:15   AdminExists() 확인
init.go:60   INSERT INTO users ...
```

**수정 방향**: `INSERT OR IGNORE` 또는 DB UNIQUE 제약으로 레벨에서 방지.

---

### 14. model.go — time.Time vs TEXT 불일치

Go struct 필드 타입이 `time.Time`인데 DB는 RFC3339 문자열 저장. 스캔 시 타입 변환 누락되면 panic 가능.

```
model.go:10, 47  CreatedAt 필드
```

실제 scan 코드에서 `string`으로 받고 있는지 확인 필요.

---

### 15. cloud.html — 다운로드 ID가 GET query string에 노출

```
cloud.html:987   /download?id=${id}
cloud.html:1049  /download?ids=${selectedIds.join(',')}
```

브라우저 히스토리, 서버 로그에 파일 ID 노출.

**수정 방향**: POST body로 전환 (발표 데모용으론 기능 문제 없음, 낮은 우선순위).

---

## Low (시간 남을 때)

### 16. cloud.html — CSRF 토큰 없음

fetch에 `credentials: 'include'`는 있지만 CSRF 토큰 없음.
서버에서 `SameSite=Lax` 쿠키 정책으로 최소 방어 가능.

### 17. share.go:375 — 에러 반환 후 쿠키 설정 코드 실행

`http.Error()` 후에도 쿠키 설정 코드가 이어지는 경로. 실제 동작엔 문제 없지만 코드 흐름 혼란.

### 18. common.go:22 — 쿠키 없음 vs DB 실패 미구분

세션 쿠키가 없는 경우와 DB 쿼리 실패를 같은 에러로 반환. 클라이언트가 원인 구분 불가.

---

## 파일별 최종 요약

| 파일 | Security | Auth | Error | DB | 종합 |
|------|----------|------|-------|----|------|
| main.go | O | O | △ | O | △ |
| db.go | O | O | △ | △ | △ |
| model.go | O | O | O | △ | O |
| user.go | O | O | △ | O | △ |
| file.go | X | O | △ | X | **X** |
| share.go | O | △ | △ | X | **X** |
| common.go | O | △ | △ | △ | △ |
| init.go | O | △ | △ | △ | △ |
| serve.go | O | O | O | O | O |
| cloud.html | △ | O | O | — | △ |
| share.html | △ | △ | O | — | △ |

(O=OK, △=주의, X=문제)

---

## 수정 우선순위 요약

1. `file.go` orphan 파일 문제 (Critical #1)
2. `share.go` tx.Commit 에러 처리 + tx.QueryRow 변경 (Critical #2, #3)
3. `user.go` JSON 디코딩 에러 처리 통일 (Critical #4)
4. `common.go` log.Fatalf → error 반환 (Critical #5)
5. `file.go` filepath.Clean + 경로 검증 (High #7)
6. 포트 상수화 (High #10)
