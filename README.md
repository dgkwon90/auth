# 인증 서비스 (Auth Service)

## 개요
이 프로젝트는 다양한 애플리케이션에서 사용할 수 있는 인증 서비스입니다. JWT 기반 로그인, 토큰 재발급, 비밀번호 재설정, 사용자 관리 기능을 제공합니다.

## 주요 기능
- 이메일/비밀번호 기반 사용자 로그인
- 로그인 시 JWT(Access/Refresh Token) 발급
- 토큰 재발급(Refresh Token)
- 비밀번호 재설정(이메일 발송)
- 사용자 정보 조회 및 수정, 탈퇴(소프트 삭제)

## 사용 기술
- Go (Golang)
- PostgreSQL
- Fiber (웹 프레임워크)
- Gmail (이메일 발송)

## 시작하기
1. 저장소 클론:
   ```
   git clone <repository-url>
   cd auth-org
   ```
2. 루트 디렉토리에 `.env` 파일을 생성하고 환경변수를 설정하세요.
3. 의존성 설치:
   ```
   go mod tidy
   ```
4. 서버 실행:
   ```
   go run cmd/main.go
   ```

## 주요 API 엔드포인트
- `POST /auth/login` : 로그인 및 JWT 발급
- `POST /auth/logout` : 로그아웃(Refresh Token 무효화)
- `POST /auth/register` : 회원가입
- `POST /auth/refresh-token` : 토큰 재발급
- `POST /auth/password/forgot` : 비밀번호 재설정 메일 발송
- `POST /auth/password/reset` : 비밀번호 재설정
- `GET /users/me` : 내 프로필 조회
- `PUT /users/me` : 내 프로필 수정
- `DELETE /users/me` : 회원 탈퇴(소프트 삭제)
- `PUT /users/me/password` : 비밀번호 변경

## API 문서(Swagger)
이 프로젝트는 [swaggo/swag](https://github.com/swaggo/swag) 및 [fiber-swagger](https://github.com/gofiber/swagger)를 사용하여 자동으로 API 문서를 생성합니다.

### API 문서 생성/업데이트 방법
1. swag CLI 설치(최초 1회):
   ```
   go install github.com/swaggo/swag/cmd/swag@latest
   ```
2. 문서 생성(프로젝트 루트에서 실행):
   ```
   swag init -g cmd/main.go
   ```
   `docs/` 디렉토리가 갱신됩니다.

### API 문서 확인 방법
- 서버 실행 후 브라우저에서 아래 주소로 접속:
  - [http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)
- 모든 엔드포인트의 요청/응답 스키마와 예시를 인터랙티브하게 확인할 수 있습니다.

## 라이선스
이 프로젝트는 MIT 라이선스를 따릅니다.