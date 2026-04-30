# TODO — 향후 작업

---

## 미완성 — 지금 당장

- [x] **로그아웃 버튼** — 유저 메뉴 dropdown으로 이동, `POST /logout` 연결 완료
- [x] **파일 확장자 필터링** — `blockedExts` map 정의, `UploadFiles` 루프 안에서 차단 확장자 skip (continue 방식)

---

## 버그

- [ ] **공유 링크 modal — 이전 링크 재표시** — 새 링크 생성 시 `shareInputRow` hidden 처리 안 됨
- [ ] **공유 링크 modal — 내용 잔류** — modal 닫혀도 제목/비밀번호/만료일 입력값 초기화 안 됨
- [ ] **다크모드 파일 색** — `EXT_COLOR` 값이 라이트모드 기준 고정색이라 다크모드에서 파일명 안 보이는 경우 있음
- [ ] **권한 체크박스 → radio 버튼 전환** — 상위 권한 선택 시 이하 권한 자동 부여, UI를 radio btn 구조로 변경 검토

---

## cloud.html — 완료

- [x] `api.getFiles()` — GET /files fetch
- [x] `api.uploadFiles()` — POST /upload fetch
- [x] `api.createShareLink()` — POST /share/create fetch (title, password, expires_hours 포함)
- [x] `api.deleteFiles(ids)` — POST /delete fetch
- [x] `api.updatePermission(fileId, permissions)` — POST /perm fetch
- [x] `api.updateOwner(fileId, targetUserId)` — POST /owner fetch
- [x] `api.getMyShareLinks()` — GET /share/list fetch
- [x] `api.getGrantedPerms()` — GET /perm/granted fetch
- [x] 파일 목록 소유자별 섹션 분리 — "내 파일 / a의 파일 / b의 파일" 섹션 헤더, 공유자 이름 가나다순
- [x] 공유 링크 목록 modal — 사이드바 "공유 링크" 버튼, 제목/파일수/만료일/잠금여부 표시, 링크 복사
- [x] 권한 부여 목록 modal — 사이드바 "권한 목록" 버튼, 파일별 그룹, 유저명 + 권한 표시
- [x] perm modal UI — 유저 검색(debounce 300ms), 드롭다운, 비트마스크 체크박스(읽기/다운로드/쓰기/삭제/관리)
- [x] owner modal UI — 유저 검색, 선택, 확인 후 소유자 변경
- [x] 삭제 확인 modal — 개별/일괄 삭제 시 "정말 삭제하시겠어요?" 확인창
- [x] 업로드 목록 스크롤 제한 — 파일 많이 추가해도 스크롤 가능, 업로드 버튼 가려지지 않음
- [x] 공유 링크 이름 말줄임 — 길면 `...` 처리 (CSS text-overflow)
- [x] 다운로드 단일/일괄 — data-action 연결, 일괄 zip
- [x] 체크박스 개별/전체 선택 — string ID 통일, indeterminate 상태
- [x] 다크모드 localStorage 지속성

## share.html — 완료

- [x] mock 데이터 제거, GET /share?token= 실제 API 연결
- [x] 비밀번호 있는 링크 — POST /share?token= 로 검증, share_session 쿠키 발급
- [x] 메타데이터 표시 — share title, 만든 사람, 만료일 (fillShareInfo)
- [x] 검색창 상단 고정 + 파일 목록 고정 높이 스크롤
- [x] 공유 파일 다운로드 — /share/download?token=&ids= 연결

## 백엔드 핸들러 — 완료

- [x] `Init` — admin 없으면 init.html, POST로 admin 계정 + 서버 이름 생성
- [x] `Login` / `Register` / `Logout`
- [x] `ListFiles` — mine/shared 분리, shared에 owner_name 포함
- [x] `UploadFiles` — 디스크 저장, file_permissions 자동 부여 (owner/admin/assiadmin)
- [x] `DownloadFiles` — 단일 스트리밍 / 복수 zip
- [x] `DeleteFiles` — PermDelete 확인, 디스크+DB 삭제
- [x] `UpdatePerm` — PermManage 확인, UPSERT
- [x] `UpdateOwner` — 소유자 변경, 권한 재조정
- [x] `CreateShareLink` — token 생성, title/password/expires 저장, 트랜잭션
- [x] `ShareInfo` (GET/POST) — 비밀번호 여부 확인, share_session 발급
- [x] `DownloadShareFiles` — share_session 검증, 소속 파일 체크, 단일/zip
- [x] `ShareHtmlServe` — /share/view → share.html 서빙
- [x] `SearchUsers` — GET /users/search?q= LIKE 쿼리, 로그인 필요
- [x] `GetMyShareLinks` — GET /share/list, 내 공유 링크 목록
- [x] `GetGrantedPerms` — GET /perm/granted, 내가 타인에게 부여한 권한 목록

---

## 중간발표 이후 (최종발표까지)

### 어드민 UI (`static/admin.html`)
- [ ] 로그인 후 Role 확인 — admin/assiadmin이면 `/admin`으로 유도
- [ ] 사용자 관리 — 목록, 역할 변경, 정지/해제, 삭제
- [ ] 파일 관리 — 전체 파일 목록(소유자 포함), 강제 삭제, 권한 수정
- [ ] 공유 링크 관리 — 전체 목록(생성자/만료일), 강제 만료
- [ ] 대시보드 — 사용자 수/파일 수/총 용량 요약
- [ ] 활성 세션 목록 + 강제 로그아웃
- [ ] 사용자별 스토리지 쿼터 설정

### 비밀번호 복구
- [ ] `users` 테이블에 `RecoveryKeyHash TEXT`, `TempKeyHash TEXT` 컬럼 추가
- [ ] 가입 완료 시 복구 키 생성 → 화면에 1회만 표시
- [ ] `handler.Recovery` / `handler.ResetPassword` 구현
- [ ] admin 복구는 assiadmin이 임시 키 발급, 일반 유저 복구는 admin이 발급

### 2. 다운로드 전송 실패 에러 처리
- 현재: `io.Copy` 시작 후 실패하면 헤더가 이미 전송된 상태라 `http.Error` 전달 안 됨
- 해결 방향: 대용량 고려해서 실시간 진행률 기능과 같이 설계 재검토

### 3. 실시간 업로드 진행률 + 이어올리기
- SSE(Server-Sent Events) 또는 WebSocket으로 클라이언트에 실시간 전달
- 청크 분할 업로드 + HTTP Range 처리

### 4. 이상 감지 및 초기화
- admin 없음, DB 손상 등 비정상 상태 → 세션 전체 무효화, 로그인으로 강제 이동

### 5. 샌드박스 보안 검사
- 위험 확장자 파일 → 임시 디렉토리에 저장 → 샌드박스 실행 → 이상 없으면 정식 저장

### 6. AI 에이전트
- 파일 분류, 위험도 판단, 이상 감지 알림 담당
- lightcloud API를 호출하는 구조로 연동

### 7. Audit Log
- `audit_logs` 테이블, 누가/언제/어떤 IP/어떤 파일 기록

### 8. File Integrity
- 업로드 시 SHA-256 해시 → DB 저장, 다운로드 시 재계산 후 비교

### 9. Rate Limiting
- 짧은 시간 내 과도한 요청 → 429 차단 (토큰 버킷, 표준 라이브러리로 구현)

### 10. 세션 바인딩
- 로그인 시 IP + User-Agent 저장, 요청마다 비교 → 다르면 무효화

---

## 최최종 (시간 남을 때)

### 11. C 연동 보안 모듈
- 암호화/복호화 로직을 C로 구현, Go에서 CGO 호출
- (시간 되면) UI 리뉴얼 — 커스텀 CSS + 3D 객체
