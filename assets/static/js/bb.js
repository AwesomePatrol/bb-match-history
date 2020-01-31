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

function fillShortMatchDetailsRows(tr, details) {
    let tbl = $("<table>")
        .append($("<thead>").append($("<tr>")
            .addClass("table-secondary")
            .append($("<td>").append("North Team [" + details.North.Players.length + "]"))
            .append($("<td>").append("South Team [" + details.South.Players.length + "]"))
        ));

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
    
    tr.append($("<td>")
        .append($("<small>").append(tbl))
        .attr("colSpan", "5")
        .attr("align", "center")
    );
}

function fillLongMatchDetailsRows(tr, details) {
    let tbl = $("<table>")
        .append($("<thead>").append($("<tr>")
            .addClass("table-secondary")
            .append($("<td>").append("Time"))
            //.append($("<td>").append("Event"))
            .append($("<td>").append("Info"))
        ));
    
    let ref = moment(details.Start);
    $.each(details.Timeline, function(index, data) {
        // Ignore some events for now.
        if (data.EventType < 4) {
            return;
        }
        let since = moment.duration(moment(data.Timestamp) - ref, 'ms');
        tbl.append($("<tr>")
            .append($("<td>").append((60*(since.hours())+since.minutes()) + ":" + since.seconds()))
            //.append($("<td>").append(data.EventType))
            .append($("<td>").append(data.Payload))
        );
    })
    
    tr.append($("<td>")
        .append($("<small>").append(tbl))
        .attr("colSpan", "5")
        .attr("align", "center")
    );
}

function getMatchDetails(event) {
    let id = event.data.ID;
    if (id == 0) {
        return;
    }
    let tr = $("<tr>")
        .attr("colSpan", "5")
        .hide();
    $(this).after(tr);
    tr.fadeIn('fast');
    let endpoint = "/api/match/short/";
    let isLong = $("#det-long").is(':checked');
    if (isLong) {
        endpoint = "/api/match/long/";
    }
    $.getJSON(endpoint + id)
        .done(function(data) {
            if (isLong) {
                fillLongMatchDetailsRows(tr, data);
            } else {
                fillShortMatchDetailsRows(tr, data);
            }
        })
        .fail(function( jqxhr, textStatus, error ) {
            var err = textStatus + ", " + error;
            console.log( "Request Failed: " + err );
            tr.append($("<td>")
                .append("Fetching recent match details failed: " + err));
        });
    $(this).one("click", {ID: id}, function(event) {
        tr.fadeOut('fast', function() {
            tr.remove();
        });
        $(this).one("click", {ID: event.data.ID}, getMatchDetails);
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
