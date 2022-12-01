Mime-Version: 1.0
From: "Team @ Dwarves Ventures" <team@dwarvesv.com>
To: {{.User.Email}}
CC: {{ccList}}
Subject: Dwarves Foundation - Allowance {{formattedCurrentMonth}} {{.Year}}
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div>
	Hi <b>{{userFirstName}}</b>,
	<br /><br />
	<div>
		Below is your payslip of {{formattedCurrentMonth}} {{.Year}}. If you have any question, please contact <a
			href="mailto:huynh@dwarvesv.com">huynh@dwarvesv.com</a>.
	</div>
	<br />
	<ul>
		<li>
			Base salary: <b>{{currency}} {{formattedBaseSalaryAmount}}</b>
		</li>
		{{if haveBonusOrCommission}}
		<li>
			Bonus:
			<ul>
				{{range projectBonusExplains}}
				<li>
					{{.Name}}: <b>{{currency}} {{.FormattedAmount}}</b>
				</li>
				{{end}}
				{{range commissionExplain}}
				<li>
					{{.Name}}: <b>{{currency}} {{.FormattedAmount}}</b>
				</li>
				{{end}}
			</ul>
		</li>
		{{end}}

		<li>Total Allowance: <b>{{currency}} {{formattedTotalAllowance}}</b></li>
		<li>TransferWise amount: <b>{{.TWAmount}} {{currencyName}}</b></li>
		<li>TransferWise conversion rate: <b>{{.TWRate}} {{currencyName}}/{{salaryCurrencyName}}</b></li>
	</ul>
	<br/><br/>
	Best regards,<br />
	<br>
	{{ template "signature.tpl" }}
</div>

--main--