var diff2str = [
  "Peaceful",
  "Piece of Cake",
  "Easy",
  "Normal",
  "Hard",
  "Nightmare",
  "Insane"
  ];

function fillMatchDetailsRows(tbl, details) {
    let n = Math.max(details.South.Players.length, details.North.Players.length);
    for (let i=0; i<n; i++) {
        let north = $("<td>");
        if (i < details.North.Players.length) {
            north.append(details.North.Players[i].Name);
        }
        let south = $("<td>");
        if (i < details.South.Players.length) {
            south.append(details.South.Players[i].Name);
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
