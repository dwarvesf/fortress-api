Mime-Version: 1.0
From: "Team @ Dwarves Ventures" <team@dwarvesv.com>
To: {{.PersonalEmail}}
Subject: Continue Using Your DF Email
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
	<div>Hi <b>{{.Name}}</b>,<div>
			<div dir=3D"ltr">
				<div><br></div>
				<div>
					We appreciate the efforts you put into your role at Dwarves Foundation. 
					While your journey with DF has concluded, we'd like to offer a small way to stay connected: you can retain your DF email alias [{{.TeamEmail}}].
				</div>
				<div><br></div>
				<div><br></div>
				<div>
					This alias will forward incoming messages to your personal email [{{.PersonalEmail}}], 
					ensuring continuity and a professional touch for your correspondence.</div><br>
					<div><br></div>
				</div>
				<div>If you have any questions, please feel free to get in touch.</div>
				<div><br></div>
				<div>Wishing you all the best,</div>
			</div>
		</div>
	</div>
	<div><br></div>-- <br>
	{{ template "signature.tpl" }}
</div>

--main--