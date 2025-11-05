Mime-Version: 1.0
From: "Team @ Dwarves Ventures" <team@d.foundation>
To: {{.PersonalEmail}}
Subject: Team Dwarves - Onboarding Succeed
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
	<div>Hi <b>{{.FullName}}</b>,<div>
			<div dir=3D"ltr">
				<div><br></div>
				<div>Your onboarding process has been succeed. You are now a {{(index .UserRank 0).Role.Name}} at
					{{.Organization}}. Please verify your Dwarves and Basecamp account, please login your work email:<br><br>
					<div>Email: <b>{{.Email}}</b><br>Password: <b>pleasechangenow</b></div><br>
					<div>You may change your password when logging in.</div>
					<div><br></div>
					<div><br></div>
				</div>
				<div>If you have any questions, please do not hesitate to contact us.</div>
				<div><br></div>
				<div>Best regards and once again, welcome to the team.</div>
			</div>
		</div>
	</div>
	<div><br></div>-- <br>
	{{ template "signature.tpl" }}
</div>

--main--