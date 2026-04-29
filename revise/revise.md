# cloud.html 작업 가이드

> 프론트 디자인 껍데기(cloud.html) 기준으로 구역별 작업 방향 정리.  
> CC와 함께 수정 작업 시 참고용.

---

## 파일 전체 구조

```
1   ~ 377줄   CSS        — 스타일 전체 (디자인 토큰, 레이아웃, 컴포넌트)
378 ~ 567줄   HTML       — 페이지 구조 (사이드바, 메인, 모달 3개)
568 ~ 1111줄  JavaScript — API 레이어, 상태 관리, 렌더링, 이벤트
1112~ 1133줄  <template> — 번들러용 SVG 썸네일 (삭제 가능)
```

---

## 🟢 건드리지 말 것 — 디자인 내부 동작

### CSS (1~377줄 전체)
다크모드 토큰, 레이아웃, 모든 컴포넌트 스타일 포함. 수정 불필요.

### JavaScript — 순수 UI 동작

| 줄 범위 | 내용 |
|---|---|
| 670~688 | `getColor()`, `formatBytes()`, `extIconHTML()` — 순수 유틸 함수 |
| 690~702 | `getFiltered()` — 검색/필터/정렬 로직 |
| 707~813 | `render()` — 파일 목록 HTML 생성 및 DOM 업데이트 전체 |
| 816~821 | `updateBulkButtons()` — 선택 개수 기반 버튼 표시 토글 |
| 844~852 | `openModal()` / `closeModal()` + 배경 클릭 / `data-close` 닫기 |
| 866~875 | 드롭존 dragover/dragleave 시각 효과 |
| 881~895 | 업로드 모달 — 파일 목록 렌더링, 항목 개별 삭제 |
| 926~931 | 공유 모달 — 권한 버튼(보기/댓글/편집) 토글 |
| 968~976 | `closeAllDropdowns()` + 바깥 클릭 시 드롭다운 닫기 |
| 1050~1057 | filter chip 클릭 → `state.mime` 변경 + `render()` |
| 1059~1067 | 테이블 헤더 클릭 → 정렬 키/방향 전환 |
| 1082~1095 | list / grid 뷰 전환 토글 |
| 1100~1105 | 다크모드 토글 |

---

## 🔴 수정 또는 삭제 대상 — 하드코딩 / Mock

### HTML

| 줄 | 내용 | 조치 |
|---|---|---|
| 411~413 | 사이드바 스토리지 `48.2 GB / 100 GB` 하드코딩 | 백엔드에서 받아 동적으로 교체 |
| 403~409 | 사이드바 폴더 nav — `All Files` 1개만 존재 | 폴더 구조 있으면 동적 생성으로 전환 |
| 456~464 | filter chip — PDF / Video / Excel / Figma / ZIP 하드코딩 | 프로젝트 파일 타입에 맞게 수정 |

### JavaScript

| 줄 | 내용 | 조치 |
|---|---|---|
| 641~648 | `EXT_COLOR` — 확장자별 색상 매핑 | 프로젝트 확장자에 맞게 수정 |
| 593 | `deleteFiles()` 내 `mockDelay()` 반환 | 실제 fetch로 교체 |
| 613 | `updatePermission()` 내 `mockDelay()` 반환 | 실제 fetch로 교체 |
| 624 | `updateOwner()` 내 `mockDelay()` 반환 | 실제 fetch로 교체 |
| 636~637 | `createShareLink()` 내 랜덤 토큰 반환 | 실제 fetch로 교체 |

### 기타

| 줄 | 내용 | 조치 |
|---|---|---|
| 1112~1130 | `<template id="__bundler_thumbnail">` | 삭제 가능 — 번들러용 SVG일 뿐 |

---

## 🟡 백엔드 연결 포인트 — 직접 수정할 부분

### `api` 객체 (576~639줄)

| 함수 | 줄 | 메서드 + 경로 | 상태 |
|---|---|---|---|
| `getFiles()` | 579~583 | `GET /files` | ✅ 실제 fetch 있음 |
| `uploadFiles()` | 597~603 | `POST /upload` | ✅ 실제 fetch 있음 |
| `deleteFiles()` | 585~594 | `DELETE /files` | ❌ mock — 주석 해제 필요 |
| `updatePermission()` | 605~614 | `PATCH /files/:id/permission` | ❌ mock — 주석 해제 필요 |
| `updateOwner()` | 616~625 | `PATCH /files/:id/owner` | ❌ mock — 주석 해제 필요 |
| `createShareLink()` | 627~638 | `POST /share` | ❌ mock — 주석 해제 필요 |

mock 교체 방법: 각 함수 내 주석 처리된 fetch 블록을 해제하고, `BASE_URL`과 `authHeaders()`를 백엔드에 맞게 정의.

```js
// 상단에 추가
const BASE_URL = 'http://localhost:8000';  // FastAPI 주소
function authHeaders() {
  return { 'Authorization': `Bearer ${your_token}` };
}
```

---

### 이벤트 핸들러 — 백엔드 연결점

| 줄 | 이벤트 트리거 | 연결 포인트 |
|---|---|---|
| 898~904 | `btnConfirmUpload` click | `api.uploadFiles()` → 완료 후 `loadFiles()` 재호출 |
| 981~1013 | `handleAction()` — `download` case | `/download?id=` 경로 백엔드에 맞게 수정 |
| 954~964 | `btnSavePerm` click | `api.updatePermission()` / `api.updateOwner()` |
| 933~936 | `btnShareSelected` click | `openShareModal()` → `api.createShareLink()` |
| 1038~1040 | `btnDeleteSelected` click | `deleteFiles()` → `api.deleteFiles()` |
| 1069~1077 | 사이드바 nav 링크 click | 현재는 `state.folder` 변경 + `render()` — 폴더별 API 호출로 확장 가능 |

---

### `loadFiles()` 데이터 매핑 (834~839줄)

백엔드 응답 필드명과 프론트 state 필드명 연결 지점.  
백엔드 응답 구조가 바뀌면 여기만 수정하면 됨.

```js
state.files = [...data.mine, ...data.shared].map(f => ({
  id: f.id,
  name: f.original_name,           // 백엔드 필드명 다르면 수정
  size: f.size,                    // 바이트 단위 숫자 or 문자열
  mime: f.mime_type.split('/')[1], // "application/pdf" → "pdf"
  uploadedTime: f.created_at.slice(0, 10) // "2025-04-28T..." → "2025-04-28"
}))
```

> ⚠️ `render()` 내부에서 참조하는 필드는 `f.uploaded`인데, 매핑에선 `uploadedTime`으로 저장됨 — 둘 중 하나로 통일 필요.

---

## 작업 순서 제안

1. `<template>` 태그 삭제 (1112~1130줄)
2. filter chip HTML을 프로젝트 파일 타입으로 교체 (456~464줄)
3. `EXT_COLOR` 수정 (641~648줄)
4. `BASE_URL` / `authHeaders()` 정의
5. mock 4개 함수 → 실제 fetch로 교체 (주석 해제)
6. `loadFiles()` 필드 매핑을 백엔드 응답에 맞게 수정 + `uploaded` / `uploadedTime` 통일
7. `download` 경로 수정 (handleAction 내 994줄)
8. 스토리지 표시 동적화 (여유 있을 때)