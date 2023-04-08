import React, {Component} from 'react';
import {EventsOff, EventsOn} from "../../wailsjs/runtime";
import {GetPeers} from "../../wailsjs/go/main/App";

export interface PeerItem {
    ip: string
    adnl: string
    uploadSpeed: string
    downloadSpeed: string
}

export interface PeersProps {
    torrent: string
}

interface State {
    peers: PeerItem[]
}

export class PeersTorrentMenu extends Component<PeersProps,State> {
    constructor(props: PeersProps, state:State) {
        super(props, state);

        this.state = {
            peers: [],
        }
    }

    update() {
        GetPeers(this.props.torrent).then((tr: any)=>{
            let newList: PeerItem[] = []
            tr.forEach((t: any)=> {
                newList.push({
                    ip: t.IP,
                    adnl: t.ADNL,
                    uploadSpeed: t.Upload,
                    downloadSpeed: t.Download,
                })
            })

            this.setState({
                peers: newList
            });
        });
    }

    componentDidMount() {
        this.update()
        EventsOn("update_peers", () => {
            this.update();
        })
    }
    componentWillUnmount() {
        EventsOff("update_peers")
    }

    renderPeersList() {
        let items = [];

        for (let t of this.state.peers) {
            items.push(<tr>
                <td>{t.ip}</td>
                <td style={{maxWidth:"200px"}}>{t.adnl}</td>
                <td className={"small"}>{t.downloadSpeed}</td>
                <td className={"small"}>{t.uploadSpeed}</td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table className="files-table" style={{fontSize:12}}>
            <thead>
            <tr>
                <th style={{width:"160px"}}>IP</th>
                <th>ADNL ID</th>
                <th style={{width:"130px"}}>Download speed</th>
                <th style={{width:"110px"}}>Upload speed</th>
            </tr>
            </thead>
            <tbody>
                {this.renderPeersList()}
            </tbody>
        </table>
    }
}