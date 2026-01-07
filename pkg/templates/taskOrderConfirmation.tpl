Mime-Version: 1.0
From: "Spawn @ Dwarves LLC" <spawn@d.foundation>
To: {{.TeamEmail}}
Subject: Monthly Task Order - {{formattedMonth}}
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div>
    <p>Hi {{contractorLastName}},</p>

    <p>This email outlines your planned assignments and work order for: <b>{{formattedMonth}}</b>.</p>

    <p>Period: <b>01 – {{periodEndDay}} {{monthName}}, {{year}}</b></p>

    <p>Active clients & locations:</p>
    <ul>
        {{range .Clients}}
        <li>{{.Name}}{{if .Country}} – {{.Country}}{{end}}</li>
        {{end}}
    </ul>

    <p>All tasks and deliverables will be tracked in Notion/Jira as usual.</p>

    <p>Please reply <b>"Confirmed – {{formattedMonth}}"</b> to acknowledge this work order and confirm your availability.</p>

    <p>Thanks,</p>

    <div><br></div>-- <br>
    {{ template "signature.tpl" }}
</div>

--main--
