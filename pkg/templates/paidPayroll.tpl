Mime-Version: 1.0
From: "Team @ Dwarves Foundation" <hr@d.foundation>
To: {{.Employee.TeamEmail}}
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
		Below is your payslip of {{formattedCurrentMonth}} {{.Year}}. If you have any questions, please open a ticket in our Discord support channel.
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

		<li>Salary Advance amount: <b>{{currency}} {{formattedSalaryAdvance}}</b></li>
		<li>Total Allowance: <b>{{currency}} {{formattedTotalAllowance}}</b></li>
		<li>TransferWise amount: <b>{{.TWAmount}} USD</b></li>
		<li>TransferWise conversion rate: <b>{{.TWRate}} USD/{{currencyName}}</b></li>
	</ul>
	<br/><br/>
	Best regards,<br />
	Dwarves Foundation <br />
	--
	<br>
	{{ template "signature.tpl" }}
</div>

--main--
