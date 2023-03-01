package notion

// MJMLTemplate is the template for the MJML email
const MJMLTemplate = `
<mjml>
  <mj-head>
    <mj-title>Changelog Email</mj-title>
    <mj-attributes>
      <mj-all font-family="Helvetica, sans-serif"></mj-all>
      <mj-section padding="0px"></mj-section>
      <mj-text font-weight="400" font-size="12px" line-height="16px" font-family="helvetica"></mj-text>
    </mj-attributes>
  </mj-head>
  <mj-body>
    <mj-section padding="20px 0">
      <mj-column>
        %s
      </mj-column>
    </mj-section>
    <mj-section>
      <mj-column>
        <mj-table>
          <tr>
            <td style="font-family: arial, helvetica, sans-serif; font-size: 12px; font-style: normal; font-weight: 400; line-height: 16px; color: #222222; max-width: 640px;">
              <p style="font-family: arial, helvetica, sans-serif;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            margin-top: 10px;
            margin-bottom: 10px;
            color: #222222;">View full archive at <a href="%s">%s</a>.</p>
              <p style="font-family: arial, helvetica, sans-serif;
            font-style: italic;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;
            margin-top: 10px;
            margin-bottom: 10px;">Copyright © 2021 Dwarves,
                LLC, All rights reserved.</p>
              <p style="font-family: arial, helvetica, sans-serif;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            margin-top: 10px;
            margin-bottom: 10px;
            color: #222222;">You're receiving this because we
                would love to have you as a part of the journey. If
                you don't want to be on the list, you can
                unsubscribe.</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 700;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">My mailing address is:</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">Hado Centrosa Garden</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">200 3/2 street</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">District 10</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">Ho Chi Minh City</p>
              <p style="font-family: Helvetica;
            font-style: normal;
            font-weight: 400;
            font-size: 11px;
            line-height: 14px;
            color: #222222;">Vietnam</p>
            </td>
          </tr>
        </mj-table>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`
