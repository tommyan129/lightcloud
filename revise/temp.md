1  # TODO — 향후 작업
        2
        3 -## 중간발표 전 — 남은 우선순위
        3 +---
        4
        5 -### 1. cloud.html fetch 연결
        6 -- [ ] `api.deleteFiles(ids)` — `POST /delete` body `{ids:[...]}
          -` 연결 (L588 비어있음)
        7 -- [ ] `api.updatePermission(fileId, permissions)` — `POST /perm
          -` body `{file_id, permissions:[{user_id, permission}]}` 연결
        8 -- [ ] `api.updateOwner(fileId, targetUser)` — `POST /owner` bod
          -y `{file_id, targetuser}` 연결
        9 -- [ ] perm modal UI — 유저 검색 input + 비트마스크 체크박스 (re
          -ad/download/write/delete/manage)
       10 -- [ ] owner modal UI — 유저 검색 input + 확인
        5 +## 미완성 — 지금 당장
        6
       12 -### 2. `/init` 초기 설정 (완료)
       13 -- [x] `handler.Init` — GET: admin 없으면 init.html 서빙, 있으면
          - `/login` 리다이렉트 / POST: 서버 이름 + admin 계정 생성
       14 -- [x] `static/init.html` — "새로운 시작" 안내 문구, 서버 이름 +
          - ID/비밀번호/재입력
       15 -- [x] `main.go` `/` 진입 시 admin 없으면 `/init` 리다이렉트
       16 -- [x] `server_settings` 테이블 추가 — `Key TEXT PK, Value TEXT
          -NOT NULL` (서버 이름 저장)
        7 +- [ ] **로그아웃 버튼** — `cloud.html` 사이드바 또는 상단에 버
          +튼 추가, `POST /logout` 연결 후 `/login` 리다이렉트
        8 +- [ ] **파일 확장자 필터링** — `UploadFiles`에서 `.exe .bat .sh
          + .ps1 .cmd` 업로드 시 400 반환
        9
       18 -### 3. 파일 확장자 필터링
       19 -- [ ] `UploadFiles`에서 `.exe .bat .sh .ps1 .cmd` 업로드 시 400
          - 반환
       20 -
       10  ---
       11
       23 -### 4. share.go — GetShareLink 구조 버그
       24 -- [ ] `GetShareLink` 맨 첫 줄 `http.ServeFile`이 응답을 완료해
          -버려서 아래 API 로직(token 파싱, DB 조회, JSON 응답) 전체가 죽
          -은 코드
       25 -- [ ] 수정 방향: 페이지 서빙(`GET /share` → HTML)과 API(`GET|PO
          -ST /share/info` → JSON) 분리, main.go 라우터도 같이 수정
       12 +## cloud.html — 완료
       13
       27 ----
       14 +- [x] `api.getFiles()` — GET /files fetch
       15 +- [x] `api.uploadFiles()` — POST /upload fetch
       16 +- [x] `api.createShareLink()` — POST /share/create fetch (title
          +, password, expires_hours 포함)
       17 +- [x] `api.deleteFiles(ids)` — POST /delete fetch
       18 +- [x] `api.updatePermission(fileId, permissions)` — POST /perm
          +fetch
       19 +- [x] `api.updateOwner(fileId, targetUserId)` — POST /owner fet
          +ch
       20 +- [x] `api.getMyShareLinks()` — GET /share/list fetch
       21 +- [x] `api.getGrantedPerms()` — GET /perm/granted fetch
       22 +- [x] 파일 목록 소유자별 섹션 분리 — "내 파일 / a의 파일 / b의
          +파일" 섹션 헤더, 공유자 이름 가나다순
       23 +- [x] 공유 링크 목록 modal — 사이드바 "공유 링크" 버튼, 제목/파
          +일수/만료일/잠금여부 표시, 링크 복사
       24 +- [x] 권한 부여 목록 modal — 사이드바 "권한 목록" 버튼, 파일별
          +그룹, 유저명 + 권한 표시
       25 +- [x] perm modal UI — 유저 검색(debounce 300ms), 드롭다운, 비트
          +마스크 체크박스(읽기/다운로드/쓰기/삭제/관리)
       26 +- [x] owner modal UI — 유저 검색, 선택, 확인 후 소유자 변경
       27 +- [x] 삭제 확인 modal — 개별/일괄 삭제 시 "정말 삭제하시겠어요?
          +" 확인창
       28 +- [x] 업로드 목록 스크롤 제한 — 파일 많이 추가해도 스크롤 가능,
          + 업로드 버튼 가려지지 않음
       29 +- [x] 공유 링크 이름 말줄임 — 길면 `...` 처리 (CSS text-overflo
          +w)
       30 +- [x] 다운로드 단일/일괄 — data-action 연결, 일괄 zip
       31 +- [x] 체크박스 개별/전체 선택 — string ID 통일, indeterminate
          +상태
       32 +- [x] 다크모드 localStorage 지속성
       33
       29 -## 중간발표 전 (완료)
       34 +## share.html — 완료
       35
       31 -### CSS 정리 (완료)
       32 -- [x] `lightcloud.css` — auth 공통 스타일 섹션 추가 (`.auth-wra
          -p`, `.auth-brand`, `.auth-card`, `.auth-field`, `.auth-input-wr
          -ap`, `.auth-submit`, `.auth-error`, `.auth-success`, `.auth-foo
          -ter`, `.auth-theme-btn`, `.field-hint`, `.pw-strength` 계열)
       33 -- [x] `login.html` / `register.html` / `recovery.html` / `init.
          -html` — 중복 스타일 제거, 각 페이지 고유 스타일만 남김
       34 -- [x] 다크 모드 `localStorage` 지속성 — 토글 시 저장, 각 페이지
          - `<head>` 복원 스크립트 추가
       36 +- [x] mock 데이터 제거, GET /share?token= 실제 API 연결
       37 +- [x] 비밀번호 있는 링크 — POST /share?token= 로 검증, share_se
          +ssion 쿠키 발급
       38 +- [x] 메타데이터 표시 — share title, 만든 사람, 만료일 (fillSha
          +reInfo)
       39 +- [x] 검색창 상단 고정 + 파일 목록 고정 높이 스크롤
       40 +- [x] 공유 파일 다운로드 — /share/download?token=&ids= 연결
       41
       36 -### share.go 완성 (완료)
       37 -- [x] `generateShareToken()` — base62 16자 토큰 생성
       38 -- [x] `CreateShareLink` — session 확인, body 파싱, 권한 확인, I
          -NSERT 완성
       39 -- [x] `getLinkFiles()` — token으로 파일 목록 조회 (내부 헬퍼)
       40 -- [x] `GetShareLink` — GET: 비밀번호 여부 확인 / POST: 비밀번호
          - 검증 후 파일 목록 반환
       41 -- [x] `DownloadShareFiles` — token 유효성 확인, map 기반 소속
          -검증, 단일 스트리밍 / 복수 zip
       42 -- [x] `main.go`에 share 라우터 연결 (`/share`, `/share/create`,
          - `/share/download`)
       42 +## 백엔드 핸들러 — 완료
       43
       44 -### main.go 라우터 (완료)
       45 -- [x] `/main` → `handler.MainPage`
       46 -- [x] `/files` → `handler.ListFiles`
       47 -- [x] `/upload` → `handler.UploadFiles`
       48 -- [x] `/download` → `handler.DownloadFiles`
       44 +- [x] `Init` — admin 없으면 init.html, POST로 admin 계정 + 서버
          + 이름 생성
       45 +- [x] `Login` / `Register` / `Logout`
       46 +- [x] `ListFiles` — mine/shared 분리, shared에 owner_name 포함
       47 +- [x] `UploadFiles` — 디스크 저장, file_permissions 자동 부여 (
          +owner/admin/assiadmin)
       48 +- [x] `DownloadFiles` — 단일 스트리밍 / 복수 zip
       49 +- [x] `DeleteFiles` — PermDelete 확인, 디스크+DB 삭제
       50 +- [x] `UpdatePerm` — PermManage 확인, UPSERT
       51 +- [x] `UpdateOwner` — 소유자 변경, 권한 재조정
       52 +- [x] `CreateShareLink` — token 생성, title/password/expires 저
          +장, 트랜잭션
       53 +- [x] `ShareInfo` (GET/POST) — 비밀번호 여부 확인, share_sessio
          +n 발급
       54 +- [x] `DownloadShareFiles` — share_session 검증, 소속 파일 체크
          +, 단일/zip
       55 +- [x] `ShareHtmlServe` — /share/view → share.html 서빙
       56 +- [x] `SearchUsers` — GET /users/search?q= LIKE 쿼리, 로그인 필
          +요
       57 +- [x] `GetMyShareLinks` — GET /share/list, 내 공유 링크 목록
       58 +- [x] `GetGrantedPerms` — GET /perm/granted, 내가 타인에게 부여
          +한 권한 목록
       59
       50 -### ListFiles 수정 (완료)
       51 -- [x] `{"mine": [...], "shared": [...]}` 두 목록 분리 반환
       52 -- [x] `model.go`에 응답용 구조체 추가 (`Mine []File`, `Shared [
          -]File`)
       53 -- [ ] UI: mine/shared 섹션 분리 표시
       60 +---
       61
       55 -### cloud.html 백엔드 연결
       56 -- [x] 폴더 사이드바 하드코딩 항목 제거 (f1~f4)
       57 -- [x] `api.getFiles()` — GET /files fetch 구현
       58 -- [x] `api.uploadFiles()` — POST /upload fetch 구현
       59 -- [x] `btnConfirmUpload` 핸들러 — MOCK_FILES 제거, loadFiles()
          -호출
       60 -- [x] 다운로드 버튼 — data-action, data-id 추가 (list + grid)
       61 -- [x] `handleAction 'download'` case 추가
       62 -- [x] `loadFiles()` — countAll 업데이트 추가
       63 -- [x] `api.createShareLink()` — POST /share fetch 구현
       64 -- [x] 필드 매핑 버그 수정 (`uploadedTime` → `uploaded`, size fo
          -rmatBytes 적용)
       65 -- [x] checkbox 개별 선택 버그 수정 — `Number(id)` → string ID로
          - 통일
       66 -- [x] 일괄 다운로드 버튼 버그 수정 — Number() 제거로 같이 해결
       67 -- [x] file size 단위 변환 — `formatBytes()` 적용, `sizeNum` 정
          -렬용 필드 분리
       68 -- [x] 업로드 중복 방지 — 버튼 클릭 시 modal 즉시 닫고 fetch 시
          -작
       69 -- [ ] `api.deleteFiles()` — 백엔드 완성, 프론트 연결 필요
       70 -- [ ] `api.updatePermission()` — 백엔드 완성, 프론트 연결 필요
       71 -- [ ] `api.updateOwner()` — 백엔드 완성, 프론트 연결 필요
       72 -- [ ] perm modal UI — 유저 검색(combobox), 다중 선택, 권한 부여
          - 연결
       73 -- [ ] owner modal UI — 유저 검색, 선택, 확인 연결
       62 +## 중간발표 이후 (최종발표까지)
       63
       75 -### /share UI (`static/share.html`)
       76 -- [ ] token으로 공유 파일 목록 표시
       77 -- [ ] 비밀번호 있는 링크면 입력 폼 먼저
       78 -- [ ] 공유 링크 생성 + QR 코드 표시 (qrcode.js 로컬)
       79 -- [ ] 개별 파일 다운로드
       80 -- [ ] 공유 모달 초기화 — 새 링크 생성 시도 시 이전 링크 결과 초
          -기화
       81 -- [ ] 내가 생성한 공유 링크 목록 조회 기능
       82 -- [ ] 검색창 상단 고정 — 파일 목록 스크롤해도 검색창 따라 내려
          -오지 않도록
       83 -- [ ] file filter chip — cloud.html의 chip 컴포넌트 그대로 이식
       84 -- [ ] 파일 목록 영역 고정 높이 + 스크롤 — 파일 수 줄어도 목록
          -영역 크기 유지
       85 -- [ ] 생성했던 공유 링크 목록 확인 가능 기능
       86 -파일 삭제 이후 링크 삭제 안됨 + 링크 목록에서도 삭제 안됨
       87 -
       88 -### cloud.html
       89 -- [ ] 공유 모달 초기화 — 새 링크 생성 시 이전 링크 결과 초기화
       90 -- [ ] 공유 모달 파일명 말줄임 — 일정 길이 초과 시 `...` 처리 (C
          -SS `text-overflow: ellipsis`)
       91 -
       92 -### 비밀번호 찾기
       93 -- [ ] `users` 테이블에 `RecoveryKeyHash TEXT`, `TempKeyHash TEX
          -T` 컬럼 추가
       94 -- [ ] 가입 완료 시 복구 키 생성 → 화면에 1회만 표시 ("지금 저장
          -하지 않으면 다시 볼 수 없습니다")
       95 -- [ ] `static/recovery.html` — 복구 키 입력 폼 + 하단 "키를 모
          -르시나요?" 안내
       96 -- [ ] `handler.Recovery` — 복구 키 bcrypt 검증 → 새 비밀번호 입
          -력 화면으로
       97 -- [ ] `handler.ResetPassword` — 새 비밀번호 받아서 hash 후 DB
          -업데이트
       98 -- [ ] login.html에 "비밀번호를 잊으셨나요?" 링크 → `/recovery`
       99 -- [ ] 어드민 패널에서 1회 한정 임시 키 발급 — 발급 즉시 `TempKe
          -yHash` 저장, 사용 후 NULL로 초기화
      100 -- [ ] admin 복구는 assiadmin이 임시 키 발급, 일반 유저 복구는 a
          -dmin이 발급
      101 -
      102 -### 초기 설정 (`/init`)
      103 -- [ ] `handler.Init` — DB에 admin 없을 때만 접근 허용, 있으면 `
          -/login` 리다이렉트
      104 -- [ ] `static/init.html` — 서버 이름 + admin ID/비밀번호/재입력
          - 입력 폼
      105 -- [ ] 안내 문구: "새로운 시작 / 관리자 계정을 설정하고 나만의
          -클라우드를 시작하세요. 한 번만 나타나는 화면입니다."
      106 -- [ ] 서버 이름 DB 저장 — `settings` 테이블 (`Key TEXT PK`, `Va
          -lue TEXT`)
      107 -- [ ] `/` 진입 시 admin 없으면 `/init`으로 리다이렉트
      108 -
       64  ### 어드민 UI (`static/admin.html`)
       65  - [ ] 로그인 후 Role 확인 — admin/assiadmin이면 `/admin`으로 유
           도
       66  - [ ] 사용자 관리 — 목록, 역할 변경, 정지/해제, 삭제
       67  - [ ] 파일 관리 — 전체 파일 목록(소유자 포함), 강제 삭제, 권한
           수정
       68  - [ ] 공유 링크 관리 — 전체 목록(생성자/만료일), 강제 만료
       69  - [ ] 대시보드 — 사용자 수/파일 수/총 용량 요약
      115 -- [ ] 설정 — 서버 이름 변경, 최대 업로드 크기, 차단 확장자 목록
          - 관리
      116 -- [ ] 사용자별 스토리지 쿼터 설정
       70  - [ ] 활성 세션 목록 + 강제 로그아웃
       71 +- [ ] 사용자별 스토리지 쿼터 설정
       72
      119 -### 백엔드 미완성
      120 -- [x] `DeleteFiles` 핸들러 — 권한 체크, 디스크+DB 삭제
      121 -- [x] `UpdatePerm` 핸들러 — PermManage 확인, UPSERT
      122 -- [x] `UpdateOwner` 핸들러 — 소유자 변경, 권한 재조정, admin 분
          -기
      123 -- [ ] 파일 확장자 필터링 — `.exe`, `.bat`, `.sh`, `.ps1` 업로드
          - 차단
       73 +### 비밀번호 복구
       74 +- [ ] `users` 테이블에 `RecoveryKeyHash TEXT`, `TempKeyHash TEX
          +T` 컬럼 추가
       75 +- [ ] 가입 완료 시 복구 키 생성 → 화면에 1회만 표시
       76 +- [ ] `handler.Recovery` / `handler.ResetPassword` 구현
       77 +- [ ] admin 복구는 assiadmin이 임시 키 발급, 일반 유저 복구는 a
          +dmin이 발급
       78
      125 ----
       79 +### 2. 다운로드 전송 실패 에러 처리
       80 +- 현재: `io.Copy` 시작 후 실패하면 헤더가 이미 전송된 상태라 `h
          +ttp.Error` 전달 안 됨
       81 +- 해결 방향: 대용량 고려해서 실시간 진행률 기능과 같이 설계 재
          +검토
       82
      127 -## 중간발표 이후 (최종발표까지)
      128 -
      129 -### 2. 다운로드 전송 실패 에러 처리 개선
      130 -
      131 -- 현재: `io.Copy` 시작 후 실패하면 헤더가 이미 전송된 상태라 `h
          -ttp.Error` 클라이언트에 전달 안 됨
      132 -- 해결 방향: `io.ReadAll`로 파일 전체를 메모리에 읽은 뒤 성공
          -시에만 헤더 + `w.Write` 순서로 응답
      133 -- 단점: 파일 전체가 메모리에 올라가므로 대용량 파일 주의 — 실시
          -간 진행률 기능과 같이 설계 재검토
      134 -
       83  ### 3. 실시간 업로드 진행률 + 이어올리기
      136 -
      137 -- `io.Copy`의 반환값(written bytes)을 활용해 진행률 계산
       84  - SSE(Server-Sent Events) 또는 WebSocket으로 클라이언트에 실시
           간 전달
      139 -- 업로드 중단 후 재개: 청크 분할 업로드 + HTTP Range 처리 필요
       85 +- 청크 분할 업로드 + HTTP Range 처리
       86
       87  ### 4. 이상 감지 및 초기화
       88 +- admin 없음, DB 손상 등 비정상 상태 → 세션 전체 무효화, 로그인
          +으로 강제 이동
       89
      143 -- admin 계정 없음, DB 손상 등 비정상 상태 감지 시 공격 또는 치
          -명적 오류로 간주
      144 -- 감지 시 처리 흐름:
      145 -  1. 서버 로그에 경고 기록
      146 -  2. 진행 중인 세션 전체 무효화
      147 -  3. 클라이언트를 초기 화면(로그인 페이지)으로 강제 이동
      148 -- 감지 시점: 서버 시작 시 + 요청 처리 중 (`UploadFiles`의 `admi
          -nID == ""` 등)
      149 -
       90  ### 5. 샌드박스 보안 검사
       91 +- 위험 확장자 파일 → 임시 디렉토리에 저장 → 샌드박스 실행 → 이
          +상 없으면 정식 저장
       92
      152 -- 위험 확장자 파일(exe, bat 등)만 대상
      153 -- 처리 흐름:
      154 -  1. 임시 디렉토리에 파일 저장
      155 -  2. 샌드박스 환경(격리된 컨테이너 또는 프로세스)에서 실행
      156 -  3. 이상 없으면 정식 저장 경로로 이동 + DB에 `locked` 상태로
          -저장
      157 -  4. 이상 감지 시 삭제 + 로그 기록
      158 -- `locked` 상태 파일은 별도 권한 없이 수정/삭제 불가
      159 -
       93  ### 6. AI 에이전트
       94 +- 파일 분류, 위험도 판단, 이상 감지 알림 담당
       95 +- lightcloud API를 호출하는 구조로 연동
       96
      162 -- 역할: 파일 관리 + 샌드박스/보안 관리 자동화
      163 -- 파일 분류, 위험도 판단, 이상 감지 알림 등 담당
      164 -- lightcloud API를 호출하는 구조로 연동 (openclaw 또는 자체 구
          -현 검토)
       97 +### 7. Audit Log
       98 +- `audit_logs` 테이블, 누가/언제/어떤 IP/어떤 파일 기록
       99
      166 -### 7. Audit Log (감사 로그)
      167 -- 누가, 언제, 어떤 IP로 어떤 파일에 접근했는지 DB에 기록
      168 -- AI 에이전트의 이상 탐지 데이터로 활용 가능
      169 -- `audit_logs` 테이블 추가, handler마다 INSERT 한 줄
      100 +### 8. File Integrity
      101 +- 업로드 시 SHA-256 해시 → DB 저장, 다운로드 시 재계산 후 비교
      102
      168 -- AI 에이전트의 이상 탐지 데이터로 활용 가능
      169 -- `audit_logs` 테이블 추가, handler마다 INSERT 한 줄
      100 +### 8. File Integrity
      101 +- 업로드 시 SHA-256 해시 → DB 저장, 다운로드 시 재계산 후 비교
      102
      171 -### 8. File Integrity (파일 무결성 검증)
      172 -- 업로드 시 SHA-256 해시 계산 → DB 저장
      173 -- 다운로드 시 또는 주기적으로 해시 재계산 → 원본과 비교
      174 -- `crypto/sha256` 표준 라이브러리로 구현 가능
      103 +### 9. Rate Limiting
      104 +- 짧은 시간 내 과도한 요청 → 429 차단 (토큰 버킷, 표준 라이브러리로 구현)
      105
      176 -### 9. Rate Limiting (요청 제한)
      177 -- 짧은 시간 내 과도한 요청 → HTTP 429 차단
      178 -- Brute Force / 가벼운 DoS 1차 방어
      179 -- `sync.Mutex` + `map`으로 토큰 버킷 직접 구현 가능
      106 +### 10. 세션 바인딩
      107 +- 로그인 시 IP + User-Agent 저장, 요청마다 비교 → 다르면 무효화
      108
      181 -### 10. 세션 바인딩 (Session Hijacking 방어)
      182 -- 로그인 시 IP + User-Agent를 세션 테이블에 저장
      183 -- 요청마다 현재 값과 비교 → 다르면 세션 무효화
      184 -
      109  ---
      110
      111  ## 최최종 (시간 남을 때)
      112
      113  ### 11. C 연동 보안 모듈
      190 -- 데이터 송수신 시 암호화/복호화 로직을 C로 구현
      191 -- Go에서 CGO로 호출하는 구조
      192 -- 목적: 표준 라이브러리 TLS 위에 추가 암호화 레이어 또는 커스텀 보안 처리
      193 -- (시간 되면) UI 리뉴얼 — pico 대신 커스텀 CSS + 3D 객체
      194 -
      114 +- 암호화/복호화 로직을 C로 구현, Go에서 CGO 호출
      115 +- (시간 되면) UI 리뉴얼 — 커스텀 CSS + 3D 객체