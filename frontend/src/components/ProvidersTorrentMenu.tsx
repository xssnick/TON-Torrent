import React, {Component} from 'react';
import {EventsOff, EventsOn} from "../../wailsjs/runtime";
import {GetPeers} from "../../wailsjs/go/main/App";
import {TonConnectButton, TonConnectUI} from "@tonconnect/ui-react";

export interface ProviderItem {
    ip: string
    adnl: string
    uploadSpeed: string
    downloadSpeed: string
}

export interface ProvidersProps {
    torrent: string
}

interface State {
    providers: ProviderItem[]
}

export class ProvidersTorrentMenu extends Component<ProvidersProps,State> {
    constructor(props: ProvidersProps, state:State) {
        super(props, state);

        this.state = {
            providers: [],
        }
    }

    update() {
       /* GetProviders(this.props.torrent).then((tr: any)=>{
            let newList: ProviderItem[] = []
            tr.forEach((t: any)=> {
                newList.push({
                    ip: t.IP,
                    adnl: t.ADNL,
                    uploadSpeed: t.Upload,
                    downloadSpeed: t.Download,
                })
            })

            this.setState({
                providers: newList
            });
        });*/
    }

    componentDidMount() {
        this.update()
        EventsOn("update_providers", () => {
            this.update();
        })
    }
    componentWillUnmount() {
        EventsOff("update_providers")

        TonConnectUI.getWallets().then((walletsList) => {
            console.log(walletsList);
        });
    }

    renderPeersList() {
        let items = [];

        for (let t of this.state.providers) {
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
        return <div>
            <TonConnectButton/>
        </div>
    }
}