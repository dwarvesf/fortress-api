Mime-Version: 1.0
From: "Spawn @ Dwarves LLC" <spawn@d.foundation>
To: {{.TeamEmail}}
Subject: Quick update for {{formattedMonth}} - Invoice reminder & client milestones
Content-Type: multipart/mixed; boundary=main

--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div>
    <p>Hi {{contractorLastName}},</p>

    <p>Hope you're having a great start to {{formattedMonth}}!</p>

    <p>Just a quick note:</p>

    <p>Your regular monthly invoice for {{formattedMonth}} services is due by <b>{{invoiceDueDay}}</b>. As usual, please use the standard template and send to <a href="mailto:billing@d.foundation">billing@d.foundation</a>.</p>

    <p>Upcoming client milestones (for awareness):</p>
    <ul>
        {{range .Milestones}}
        <li>{{.}}</li>
        {{end}}
    </ul>

    <p>You're continuing to do excellent work on the embedded team – clients are very happy with your contributions.</p>

    <p>If anything comes up or you need support, just ping me anytime.</p>

    <p>Best,</p>

    <div><br></div>-- <br>
    {{ template "signature.tpl" }}
</div>

--main--
