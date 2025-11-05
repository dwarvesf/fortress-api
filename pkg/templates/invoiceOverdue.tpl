Mime-Version: 1.0
From: "Accounting @ Dwarves Ventures" <accounting@d.foundation>
To: {{.Email}}
CC: {{gatherAddress}}
References: {{.MessageID}}
In-Reply-To: {{.MessageID}}
Subject: Dwarves Foundation - Invoice #{{.Number}} for {{toString .Month}} {{.Year}} work on {{.Project.Name}}
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
	<div>Hi {{.Project.Name}} team,<div>
			<div dir=3D"ltr">
				<div><br></div>
				<div>Did you have a chance to complete the payment for the Invoice #{{.Number}} ? We haven't received it yet.
				</div>
				<div><br></div>
				<div><br></div>
			</div>
			<div dir=3D"ltr">
				<div>Thank you,</div>
			</div>
		</div>
	</div>
	<div><br></div>-- <br>
	{{ template "signature.tpl" }}
</div>

--main--