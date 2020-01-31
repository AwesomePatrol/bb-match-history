var diff2str = [
  "Peaceful",
  "Piece of Cake",
  "Easy",
  "Normal",
  "Hard",
  "Nightmare",
  "Insane"
  ];

function showDifficultyBreakdown() {
    let df = $("#diffBreak");
    df.empty();
    let tbl = $("<table>")
        .addClass("table")
        .addClass("table-hover")
        .addClass("table-striped");
    let th = $("<tr>");
    $.each(diff2str, function(i, val) {
        th.append($("<td>").append(val));
    });
    tbl.append($("<thead>")
        .append(th)
        .addClass("table-secondary")
    );
    let tr = $("<tr>");
    $.each(diff2str, function(i, val) {
        tr.append($("<td>").append($("#matches td:contains(" + val + ")").length));
    });
    tbl.append($("<tbody>").append(tr));
    df.append(tbl);
    df.show();
}

function fillMatchDetailsRows(tbl, details) {
    let n = Math.max(details.South.Players.length, details.North.Players.length);
    for (let i=0; i<n; i++) {
        let north = $("<td>");
        if (i < details.North.Players.length) {
            let name = details.North.Players[i].Name;
            north.append($("<a>")
                .attr("href", "/site/search/?name=" + encodeURIComponent(name))
                .append(name));
        }
        let south = $("<td>");
        if (i < details.South.Players.length) {
            let name = details.South.Players[i].Name;
            south.append($("<a>")
                .attr("href", "/site/search/?name=" + encodeURIComponent(name))
                .append(name));
        }
        tbl.append($("<tr>")
            .append(north)
            .append(south)
        );
    }
}

function getMatchDetails(event) {
    let id = event.data.ID;
    if (id == 0) {
        return;
    }
    let tr = $("<tr>")
        .attr("colSpan", "5");
    $(this).after(tr);
    $.getJSON( "/api/match/short/" + id)
        .done(function(data) {
            let tbl = $("<table>")
                .append($("<thead>").append($("<tr>")
                    .addClass("table-secondary")
                    .append($("<td>").append("North Team [" + data.North.Players.length + "]"))
                    .append($("<td>").append("South Team [" + data.South.Players.length + "]"))
                ));
            fillMatchDetailsRows(tbl, data);
            tr.append($("<td>")
                .append($("<small>").append(tbl))
                .attr("colSpan", "5")
                .attr("align", "center")
            );
        })
        .fail(function( jqxhr, textStatus, error ) {
            var err = textStatus + ", " + error;
            console.log( "Request Failed: " + err );
            tr.append($("<td>")
                .append("Fetching recent match details failed: " + err));
        });
}

function addRecentMatchesEntry(tbl, match) {
    let ago = moment(match.Start);
    let winner = "South";
    if (match.NorthWon) {
      winner = "North";
    }
    let diffStr = diff2str[match.Difficulty];
    let row = $("<tr>")
      .append($("<td>").append(match.ID))
      .append($("<td>").append(ago.fromNow()))
      .append($("<td>").append(moment.duration(match.Length/1000000).humanize()))
      .append($("<td>").append(winner))
      .append($("<td>").append(diffStr))
      .one("click", {ID: match.ID}, getMatchDetails);
    if (!(typeof match.IsWinner === 'undefined')) {
        if (match.IsWinner) {
            row.addClass("table-success")
        } else {
            row.addClass("table-danger")
        }
    }
    tbl.append(row);
}
