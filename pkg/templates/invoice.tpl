Mime-Version: 1.0
From: "Accounting @ Dwarves Ventures" <accounting@dwarvesv.com>
To: {{.Email}}
CC: {{gatherAddress}}
Subject: Dwarves Foundation - Invoice #{{.Number}} for {{toString .Month}} {{.Year}} work on {{.Project.Name}}
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
	<div>Hi {{.Project.Name}} team,<div>
			<div dir=3D"ltr">
				<div><br></div>
				<div>Thank you for working with us. We always wish for your success. Throughout the last month, do you have any
					feedback on our work to better our collaboration?</div>
				<div><br></div>
				<div>Should our work meet your expectation, you could find our invoice attached to this email and complete the
					payment<br></div>
				<div><br></div>
			</div>
			<blockquote style=3D"margin:0px 0px 0px 40px;border:none;padding:0px">
				<div>Invoice number: #{{.Number}}
				</div>
				<div>
					<div>Project: {{.Project.Name}}.</div>
					<div>{{description}}</div>
				</div>
				<div>
					Invoice Date: {{formatDate .InvoiceDate}}.</div>
				<div>Due Date: {{formatDate .DueDate}}.
				</div>
				<div><br></div>
				<div><br></div>
			</blockquote>
			<div dir=3D"ltr">
				<div>If you have any questions, please do not hesitate to contact =
					us.</div>
				<div><br></div>
				<div>Best regards,</div>
			</div>
		</div>
	</div>
	<div><br></div>-- <br>
	{{ template "signature.tpl" }}
</div>

--main
Content-Type: application/pdf; name="{{.Number}}.pdf"
Content-Disposition: attachment; filename="{{.Number}}.pdf"
Content-Transfer-Encoding: base64

{{encodedPDF}}

--main--