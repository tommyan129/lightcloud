# lightcloud - 중간 발표 전 수정 및 개선 사항 (beforeMid2.md)

`@checklist.md`를 기준으로 현재 코드베이스의 문제점과 수정해야 할 점을 정리했습니다.

## 1. 보안 (Security)
- **MIME 타입 검증 강화**: `handler/file.go`의 `UploadFiles`에서 파일 타입을 클라이언트가 보낸 `Content-Type` 헤더에 의존하고 있습니다. `http.DetectContentType`을 사용하여 실제 파일 내용으로 검증해야 합니다.
- **Path Traversal 방지**: 파일 업로드/다운로드 경로 결합 시 `filepath.Clean()`을 명시적으로 사용하여 경로 탈출 공격을 방지하세요.
- **파일 크기 제한**: `r.ParseMultipartForm`으로 버퍼 크기는 제한하지만, 전체 request body 크기를 `http.MaxBytesReader` 등을 통해 제한하는 로직이 부족합니다.

## 2. 인증 및 세션 (Authentication & Session)
- **세션 검증 일관성**: 대부분의 핸들러에서 `getSessionUser`를 호출하지만, 호출하지 않는 핸들러(예: `MainPage` 등)가 있는지 다시 한번 확인이 필요합니다.
- **로그아웃 로직**: `Logout` 핸들러에서 DB 삭제 실패 시에도 쿠키 만료 처리를 하는 등 흐름을 더 견고하게 다듬을 필요가 있습니다.

## 3. 에러 핸들링 (Error Handling)
- **에러 메시지 형식 통일**: 어떤 곳은 `http.Error`를 쓰고, 어떤 곳은 `w.WriteHeader`와 JSON 응답을 직접 작성합니다. 프로젝트 전체의 에러 응답 형식을 하나로 통일하세요.
- **로깅 보완**: 핸들러 내부에서 에러 발생 시 `log.Printf`로 남기는 메시지에 더 구체적인 컨텍스트(사용자 ID, 파일 ID 등)를 포함하세요.

## 4. 코드 구조 (Code Structure)
- **[중요] DB 로직 분리**: 현재 모든 DB 쿼리가 `handler/` 패키지의 함수 내부에 직접 작성되어 있습니다. 체크리스트 권장사항에 따라 DB CRUD 로직은 `db/` 패키지(예: `db/user_repo.go`, `db/file_repo.go`)로 이동시켜야 합니다.
- **중복 코드 제거**: 파일 권한 확인(`p & model.PermDownload == 0`)과 같은 로직이 여러 곳에 흩어져 있습니다. `common.go` 등에 유틸리티 함수로 추출하세요.

## 5. HTTP 및 라우팅 (HTTP & Routing)
- **HTTP Method 검증**: `ListFiles`, `UploadFiles` 등의 핸들러에서 `r.Method`를 명시적으로 체크하지 않는 경우가 있습니다. 의도하지 않은 Method 요청에 대해 `405 Method Not Allowed`를 반환하도록 보완하세요.
- **하드코딩 제거**: `main.go`의 포트 번호(`:8080`), `file.go`의 업로드 경로(`./upload`) 등을 상수로 정의하거나 설정 파일로 분리하세요.

## 6. 데이터베이스 (Database)
- **에러 처리 누락**: `UpdateOwner` 핸들러 등 일부 코드에서 에러가 발생해도 주석(`//err`)만 있고 처리가 누락된 부분이 있습니다. 모든 에러는 적절히 처리하거나 로그를 남겨야 합니다.

## 7. 안정성 (Stability)
- **Nil 포인터 및 예외 처리**: DB 조회 결과가 없을 때(`sql.ErrNoRows`)의 처리가 모든 곳에서 완벽한지 확인하세요. 특히 다중 파일 처리 시 하나라도 실패할 경우의 롤백 처리를 점검해야 합니다.
