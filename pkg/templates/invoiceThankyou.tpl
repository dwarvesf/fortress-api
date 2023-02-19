Mime-Version: 1.0
From: "Accounting @ Dwarves Ventures" <accounting@dwarvesv.com>
To: {{.Email}}
CC: {{gatherAddress}}
References: {{.References}}
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
				<div>Thanks for completing the payment for Invoice #{{.Number}}. This is to confirm that we have received your
					payment fully.</div>
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