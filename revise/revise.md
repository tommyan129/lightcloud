# lightcloud 수정 및 개선 사항 리스트 (Revise List)

## 1. 치명적 버그 (Critical Bugs)
- [ ] **로그아웃 라우팅 누락**: `main.go`에 `handler.Logout` 라우팅 등록 필요. 현재 유저가 로그아웃할 방법이 없음.
- [ ] **자원 해제(Close) 미흡**: `handler/file.go` 및 `handler/share.go` 내 파일 열기(`os.Open`), ZIP 생성(`zip.NewWriter`) 시 `defer`를 사용하지 않아 에러 발생 시 자원 누수 및 파일 깨짐 위험이 있음.
- [ ] **트랜잭션 미적용**: `UploadFiles` 핸들러에서 파일 저장과 DB 입력을 별도로 수행함. 중간에 실패할 경우 파일만 남거나 DB 데이터만 남는 불일치 발생 가능. `db.Begin()` 트랜잭션 적용 필요.
- [ ] **공유 링크 다운로드 보안 우회**: `DownloadShareFiles` 핸들러에서 토큰 만료 여부만 확인하고 비밀번호 검증을 수행하지 않음. 비밀번호가 설정된 링크라도 토큰만 알면 직접 다운로드 가능.

## 2. 프론트엔드 (cloud.html) 필수 수정
- [ ] **ID 타입 변환 제거**: JS 내의 `Number(id)` 강제 형변환 로직 전부 삭제 (UUID 문자열 그대로 사용).
- [ ] **handleAction 변수 오류**: `numId`, `id` 등 정의되지 않은 변수 참조 오류 해결 및 파라미터 전달 교정.
- [ ] **하드코딩 UI 동적화**: 용량 표시바(48.2 GB 등)를 실제 DB 사용량 및 설정된 할당량(Quota) 기반으로 표시하도록 수정.
- [ ] **API 객체 완성**: `api.deleteFiles`, `api.updatePermission` 등 비어있는 함수들을 백엔드 핸들러와 연결.

## 3. 기능 확장 및 연동 (Backend + Frontend)
- [ ] **폴더 네비게이션 구현**: 현재 UI만 있는 사이드바의 폴더 기능을 백엔드 DB(files 테이블에 Folder 컬럼 추가 등)와 연동.
- [ ] **권한/소유자 수정 시스템**:
    - 백엔드: 권한 수정(Bitmask 적용) 및 소유자 변경 핸들러 작성.
    - 프론트: 현재의 단순 Select 방식을 비트마스크 설정이 가능한 UI(체크박스 등)로 재설계 및 모달 연동.
- [ ] **어드민 권한 로직 개선**: 파일 업로드 시 모든 어드민(`Role='admin'`)에게 권한을 부여하도록 `Query` 로직으로 변경.

## 4. 구조 개선 및 기타 (Refactoring)
- [ ] **데이터 타입 일관성**: `model.File`의 `CreatedAt`을 포함한 모든 시간 데이터를 RFC3339 `string`으로 통일하여 SQLite 호환성 확보.
- [ ] **중복 코드 통합**: `DownloadFiles`와 `DownloadShareFiles`의 파일 전송/ZIP 생성 로직 유틸리티화.
- [ ] **서버 시작 로그**: `main.go`의 `http.ListenAndServe` 실패 시 `log.Fatal`로 로그 남기기.
- [ ] **DB 외래 키 보강**: `share_links(CreatedBy)` 등 누락된 외래 키 제약 조건 추가.

---
- ## 4. 기타 개선 사항 (Etc)
      19 - - [ ] **DB 외래 키 누락**: `share_links` 테이블의 `CreatedBy` 컬럼에 `users(ID)` 외래 키 제약 조건 추가 필요.
      20 - - [ ] **서버 시작 로그**: `main.go`의 `http.ListenAndServe` 리턴값을 체크하여 서버 시작 실패 시 즉시 로그를 남기고 종료하도록 수정.
      21 - - [ ] **cloud.html (Front-end)**: `CLAUDE.md`에 명시된 대로 불필요한 mock 코드를 제거하고 실제 백엔드 API와 연결하는 작업 필요.
       7 -
       8 - ## 2. 보안 취약점 (Security Vulnerabilities)
       7   - [ ] **공유 링크 다운로드 보안 우회**: `DownloadShareFiles` 핸들러에서 토큰 만료 여부만 확인하고 비밀번호 검증을 수행하지 않음. 비밀번호가 설정된 링크라도 토큰만 알면 직접
           다운로드 가능.
      10 - - [ ] **어드민 권한 부여 로직**: 파일 업로드 시 `QueryRow`로 어드민 1명에게만 권한을 부여함. 어드민이 여러 명일 경우를 대비해 `Query`를 이용해 모든 어드민에게 권한을 주도록
         수정 필요.
      11 - - [ ] **Login JSON 디코딩 에러 처리**: `Login` 핸들러에서 `json.NewDecoder` 결과에 대한 에러 처리가 누락되어 잘못된 요청 시 서버 동작이 불안정할 수 있음.
       8
      13 - ## 3. 코드 구조 및 리팩토링 (Refactoring)
      14 - - [ ] **중복 코드 통합**: `DownloadFiles`와 `DownloadShareFiles`의 파일 전송/ZIP 생성 로직이 거의 동일함. 이를 별도의 유틸리티 함수나 헬퍼로 분리하여 코드 중복 제거.
      15 - - [ ] **세션 헬퍼 고도화**: `getSessionUser`가 유저 ID만 반환하는데, 권한 체크를 위해 `*model.User` 객체 전체를 반환하도록 고도화하고 위치를 `handler/common.go` 등으로
         공통화 제안.
      16 - - [ ] **데이터 타입 일관성**: `model.File`의 `CreatedAt`만 `string`으로 되어 있음. 다른 모델들처럼 `time.Time`으로 통일하여 시간 연산 및 정렬 편의성 확보.