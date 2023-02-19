Mime-Version: 1.0
From: "Dwarvesv Team" <team@dwarvesv.com>
To: spawn@dwarvesv.com
Subject: Apply job {{.Role}} from {{.Email}}.
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
	<div>
    <div>
			<blockquote style=3D"margin:0px 0px 0px 40px;border:none;padding:0px">
				<div>Name: {{.Name}}</div>
				<div>Email: {{.Email}}</div>
				<div>Detail:</div>
				<div><p style=3D"white-space: pre-wrap;">{{.Detail}}</p></div>
			</blockquote>
		</div>
	</div>
	<div><br></div>-- <br>
	{{ template "signature.tpl" }}
</div>

--main
Content-Type: application/pdf; name="{{.Name}}.pdf"
Content-Disposition: attachment; filename="{{.Name}}.pdf"
Content-Transfer-Encoding: base64

{{encodedPDF}}

--main--
