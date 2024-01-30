import React, { useState, useEffect } from 'react';
import { EventsOff, EventsOn } from "../../wailsjs/runtime";
import { GetPeers } from "../../wailsjs/go/main/App";

export interface PeerItem {
    ip: string;
    adnl: string;
    uploadSpeed: string;
    downloadSpeed: string;
}

export interface PeersProps {
    torrent: string;
}

interface State {
    peers: PeerItem[];
}

const PeersTorrentMenu: React.FC<PeersProps> = (props) => {
    const [state, setState] = useState<State>({
        peers: [],
    });

    const update = () => {
        GetPeers(props.torrent).then((tr: any) => {
            let newList: PeerItem[] = [];
            tr.forEach((t: any) => {
                newList.push({
                    ip: t.IP,
                    adnl: t.ADNL,
                    uploadSpeed: t.Upload,
                    downloadSpeed: t.Download,
                });
            });

            setState({
                peers: newList,
            });
        });
    };

    useEffect(() => {
        update();
        EventsOn("update_peers", () => {
            update();
        });

        return () => {
            EventsOff("update_peers");
        };
    }, [props.torrent]);

    const renderPeersList = () => {
        let items: JSX.Element[] = [];

        for (let t of state.peers) {
            items.push(
                <tr key={t.ip}>
                    <td>{t.ip}</td>
                    <td style={{ maxWidth: "200px" }}>{t.adnl}</td>
                    <td className={"small"}>{t.downloadSpeed}</td>
                    <td className={"small"}>{t.uploadSpeed}</td>
                </tr>
            );
        }
        return items;
    };

    return (
        <table className="files-table" style={{ fontSize: 12 }}>
            <thead>
            <tr>
                <th style={{ width: "160px" }}>IP</th>
                <th>ADNL ID</th>
                <th style={{ width: "130px" }}>Download speed</th>
                <th style={{ width: "110px" }}>Upload speed</th>
            </tr>
            </thead>
            <tbody>
            {renderPeersList()}
            </tbody>
        </table>
    );
};

export default PeersTorrentMenu;
