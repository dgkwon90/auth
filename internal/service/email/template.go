package email

const passwordResetEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>비밀번호 재설정 안내</title>
</head>
<body style="font-family: Arial, sans-serif; background: #f8f8f8; padding: 30px;">
  <div style="max-width: 480px; margin: auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #eee; padding: 32px;">
    <h2 style="color: #1a73e8;">비밀번호 재설정 요청</h2>
    <p>안녕하세요,</p>
    <p>비밀번호 재설정 요청을 받았습니다.<br>
      아래 버튼을 클릭하여 새로운 비밀번호를 설정하세요.</p>
    <p style="text-align: center;">
      <a href="{{.ResetLink}}" style="display:inline-block; background:#1a73e8; color:#fff; padding:12px 24px; border-radius:5px; text-decoration:none; font-weight:bold;">
        비밀번호 재설정하기
      </a>
    </p>
    <p>이 링크는 <b>{{.ExpireMinutes}}분</b> 동안만 유효합니다.<br>
      만약 본인이 요청하지 않았다면 이 메일을 무시하셔도 됩니다.</p>
    <hr style="margin:32px 0 16px 0;">
    <small style="color:#888;">본 메일은 자동 발송된 메일입니다.</small>
  </div>
</body>
</html>
`

type PasswordResetEmailData struct {
	ResetLink     string
	ExpireMinutes int
}
