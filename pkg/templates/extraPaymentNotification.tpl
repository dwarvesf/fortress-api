Mime-Version: 1.0
From: "Operations @ Dwarves LLC" <spawn@d.foundation>
To: {{.ContractorEmail}}
Subject: Extra Payment Approval - {{.Month}}
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div>
    <p>Hi {{contractorFirstName}},</p>

    <p>As a discretionary recognition of your exceptional contributions this month{{if .Reasons}}:</p>
    <ul>
        {{range .Reasons}}
        <li>{{.}}</li>
        {{end}}
    </ul>
    <p>We've{{else}}, we've{{end}} approved an extra payment of <b>{{.AmountFormatted}}</b>.</p>

    <p>Please include it as a separate line on your next invoice:</p>
    <ul>
        <li>Description: Extra Payment - {{.Month}}</li>
        <li>Amount: {{.AmountFormatted}} USD</li>
    </ul>
    <p>No other changes needed.</p>

    <p>Thanks again for the great work!</p>

    <p>Best,</p>

    <div><br></div>-- <br>
    {{ template "signature.tpl" }}
</div>

--main--
