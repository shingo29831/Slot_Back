
async function fetchLogs() {
    console.log("test")
    try {
        const response = await fetch('/api/logs');
        if (!response.ok) {
            throw new Error('データの取得に失敗しました');
        }

        const logs = await response.json();
        updateLogTable(logs);
    } catch (error) {
        console.error(error);
    }
}
function lv_int2str(i){
    switch (i) {
        case 0:return "succsess"
        case 1:return "note"
        case 2:return "warning"
        case 3:return "error"
        default :return "????"
    }
}
function updateLogTable(logs) {
    const logTable = document.getElementById('log-table-body');
    logTable.innerHTML = ''; // テーブルをクリア

    logs.forEach(log => {
        const row = document.createElement('tr');
        // 重要度の色分け
        let levelColor;
        switch (log.level) {
            case 3:
                levelColor = 'red';
                break;
            case 2:
                levelColor = 'orange';
                break;
            default:
                levelColor = 'green';
        }
        // 各セルを作成してデータを挿入
        row.innerHTML = `
            <td style="color:${levelColor}">${lv_int2str(log.level)}</td>
            <td>${log.location}</td>
            <td>${log.message}</td>
            <td>${log.time}</td>
        `;
        tableBody.appendChild(row);
    });
    console.log("all done")
}

function main(){
    document.addEventListener('DOMContentLoaded', function() {
        
        // サーバーからログデータを取得する
        fetch('/api/logs') // サーバーのエンドポイントを指定
            .then(response => response.json())
            .then(data => {
                const logTable = document.getElementById('log-table-body');
                logTable.innerHTML = ''; // テーブルをクリア

                data.forEach(log => {
                    const row = document.createElement('tr');

                    // 重要度の色分け
                    let levelColor;
                    switch (log.level) {
                        case 3:
                            levelColor = 'red';
                            break;
                        case 2:
                            levelColor = 'orange';
                            break;
                        default:
                            levelColor = 'green';
                    }

                    // 各セルを作成してデータを挿入
                    row.innerHTML = `
                        <td style="color:${levelColor}">${lv_int2str(log.level)}</td>
                        <td>${log.location}</td>
                        <td>${log.message}</td>
                        <td>${log.time}</td>
                    `;

                    // テーブルに行を追加
                    logTable.appendChild(row);
                });
            })
            .catch(error => {
                console.error('データの取得に失敗しました:', error);
            });
        }
    );
}

main()

setInterval(main, 1000)