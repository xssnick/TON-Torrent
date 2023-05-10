import React, {MouseEvent, Component} from 'react';
import Logo from "./assets/images/logo.svg"
import Download from "./assets/images/icons/download.svg"
import './tooltip.css';
import './modal.scss';
import {Filter, Refresh, SelectedTorrent, Table} from "./components/Table";
import {AddTorrentModal} from "./components/ModalAddTorrent";
import {WaitReady, SetActive, WantRemoveTorrent} from "../wailsjs/go/main/App";
import {FiltersMenu} from "./components/FiltersMenu";
import {BrowserOpenURL, EventsOn} from "../wailsjs/runtime";
import {FilesTorrentMenu} from "./components/FilesTorrentMenu";
import {CreateTorrentModal} from "./components/ModalCreateTorrent";
import {PeersTorrentMenu} from "./components/PeersTorrentMenu";
import {InfoTorrentMenu} from "./components/InfoTorrentMenu";
import {SettingsModal} from "./components/ModalSettings";
import {RemoveConfirmModal} from "./components/ModalRemoveConfirm";

interface State {
    selectedItems:  SelectedTorrent[]
    infoSize: number
    tableFilter: Filter
    showAddTorrentModal: boolean
    showCreateTorrentModal: boolean
    showSettingsModal: boolean
    showRemoveConfirmModal: boolean

    overallUploadSpeed: string
    overallDownloadSpeed: string
    torrentMenuSelected: number

    ready: boolean

    openFileHash?: string
    removeHash?: string
}

export class App extends Component<{}, State> {
    constructor(props: any, state: State) {
        super(props, state);

        this.state = {
            selectedItems: [],
            infoSize: 150,
            tableFilter: {
                type: "all",
                search: "",
            },
            showAddTorrentModal: false,
            showCreateTorrentModal: false,
            showSettingsModal: false,
            showRemoveConfirmModal: false,
            overallUploadSpeed: "0 Bytes",
            overallDownloadSpeed: "0 Bytes",
            torrentMenuSelected: -1,
            ready: false,
        }
    }

    setSelectedTorrentMenu = (n: number) => {
        return () => {
            this.setState((current)=>({...current, torrentMenuSelected: n}))
        }
    }

    hasActiveTorrents = () => {
        let has = false;
        this.state.selectedItems.forEach((si)=> {
            if (si.active) has = true;
        })
        return has;
    }
    hasInactiveTorrents = () => {
        let has = false;
        this.state.selectedItems.forEach((si)=> {
            if (!si.active) has = true;
        })
        return has;
    }

    toggleAddTorrentModal = () => {
        this.setState((current)=>({...current, showAddTorrentModal: !this.state.showAddTorrentModal, openFileHash: undefined}))
    }
    toggleCreateTorrentModal = () => {
        this.setState((current)=>({...current, showCreateTorrentModal: !this.state.showCreateTorrentModal}))
    }
    toggleSettingsModal = () => {
        this.setState((current)=>({...current, showSettingsModal: !this.state.showSettingsModal}))
    }
    toggleRemoveConfirmModal = () => {
        this.setState((current)=>({...current, showRemoveConfirmModal: !this.state.showRemoveConfirmModal, removeHash: undefined}))
    }

    componentDidMount() {
        EventsOn("want_remove_torrent", (hash: string) => {
            this.setState((current)=>({...current, removeHash: hash, showRemoveConfirmModal: true}))
        })
        EventsOn("open_torrent", (hash: string) => {
            this.setState((current)=>({...current, showAddTorrentModal: true, openFileHash: hash}))
        })
        EventsOn("daemon_ready", (data)=> {
            this.setState((current)=>({...current, ready: true}));
        })
        WaitReady().then()

        window.addEventListener('resize', () => {
            let inf = document.getElementsByClassName("torrent-info");
            if (inf.length == 0) {
                return
            }

            let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height;
            let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;
            let minINF = inf[0].getBoundingClientRect();

            let minH = minINF.height;
            if (minH > window.innerHeight-(topH+botH)) {
                this.setState((current)=>({...current, infoSize: window.innerHeight-(topH+botH)}));
            }
        })

        EventsOn("speed", (data)=> {
            this.setState((current)=>({...current, overallUploadSpeed: data.Upload, overallDownloadSpeed: data.Download}));
        })
    }

    extendInfoEvent = (mouseDownEvent: MouseEvent) =>  {
        const startSize = this.state.infoSize;
        const startPosition = { x: mouseDownEvent.pageX, y: mouseDownEvent.pageY };

        const onMouseMove = (mouseMoveEvent: any) => {
            let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height;
            let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;

            let sz = startSize + startPosition.y - mouseMoveEvent.pageY;
            if (sz > window.innerHeight-(topH+botH)) {
                sz = window.innerHeight-(topH+botH)
            }

            if (sz < 20) {
                sz = 20
            }
            this.setState((current)=>({...current, infoSize: sz}));
        }
        const onMouseUp = () => {
            document.body.removeEventListener("mousemove", onMouseMove);
        }
        document.body.addEventListener("mousemove", onMouseMove);
        document.body.addEventListener("mouseup", onMouseUp, { once: true });
    }

    render() {
        return (
            <div id="App">
                <div className="daemon-waiter" style={this.state.ready ? {display: "none"} : {}}>
                    <div className="loader-block">
                        <span className="loader"/><span className="loader-text">Initializing storage daemon...</span>
                    </div>
                </div>
                {this.state.showAddTorrentModal ? <AddTorrentModal openHash={this.state.openFileHash} onExit={this.toggleAddTorrentModal}/> : null}
                {this.state.showCreateTorrentModal ? <CreateTorrentModal onExit={this.toggleCreateTorrentModal}/> : null}
                {this.state.showSettingsModal ? <SettingsModal onExit={this.toggleSettingsModal}/> : null}
                {this.state.showRemoveConfirmModal ? <RemoveConfirmModal hash={this.state.removeHash!}  onExit={this.toggleRemoveConfirmModal}/> : null}
                <div className="left-bar">
                    <div className="logo-block">
                        <img className="logo-img" src={Logo} alt=""/>
                        <label className="logo-text">TON Torrent</label>
                    </div>
                    <div className="menu-block">
                        <FiltersMenu onChanged={(v) => {
                            this.setState((current)=>({...current, tableFilter: {
                                    type: v,
                                    search: this.state.tableFilter.search
                                }}));
                            Refresh();
                        }}/>
                        <div className="actions-menu">
                            <button className="menu-item main" onClick={this.toggleAddTorrentModal}>
                                Add Torrent
                            </button>
                            <button className="menu-item" onClick={this.toggleCreateTorrentModal}>
                                Create Torrent
                            </button>
                            <button className="menu-item" onClick={this.toggleSettingsModal}>
                                Settings
                            </button>
                        </div>
                    </div>
                    <div className="version-block">
                        <div className="ver-info">
                            <span>v0.1.0</span>
                            <button className="updates" onClick={()=>{
                                BrowserOpenURL("https://github.com/tonutils/torrent-client/releases")
                            }}>Check updates</button>
                        </div>
                    </div>
                </div>
                <div className="right-screen">
                    <div className="top-bar">
                        <div className="top-buttons-container">
                            <button className={this.hasInactiveTorrents() ? "top-button start" : "top-button start disabled"} disabled={!this.hasInactiveTorrents()} onClick={() => {
                                this.state.selectedItems.forEach((t) => {
                                    SetActive(t.hash, true).then(Refresh)
                                })
                            }}/>
                            <button className={this.hasActiveTorrents() ? "top-button stop" : "top-button stop disabled"} disabled={!this.hasActiveTorrents()} onClick={() => {
                                this.state.selectedItems.forEach((t) => {
                                    SetActive(t.hash, false).then(Refresh)
                                })
                            }}/>
                            <button className={this.state.selectedItems.length > 0 ? "top-button remove" : "top-button remove disabled"} disabled={this.state.selectedItems.length == 0} onClick={() => {
                                this.state.selectedItems.forEach((t) => {
                                    WantRemoveTorrent(t.hash).then(()=>{
                                      //  Refresh()
                                      //  this.setState((current) => ({ ...current, selectedItems: []}))
                                    })
                                })
                            }}/>
                        </div>
                        <input type="text" className="search-input" placeholder="Search..." onChange={(e) => {
                            this.setState((current)=>({...current, tableFilter: {
                                    type: this.state.tableFilter.type,
                                    search: e.target.value,
                                }}));
                        }}/>
                    </div>
                    <div className="torrents-table" style={{height: "50%", maxWidth: '100%', overflowX: "auto"}}>
                        <Table filter={this.state.tableFilter} onSelect={(sl) => {
                            let menu = this.state.torrentMenuSelected;
                            if (menu == -1 && sl.length > 0) {
                                menu = 0;
                            } else if (sl.length == 0) {
                                menu = -1;
                            }

                            this.setState((current) => ({ ...current, selectedItems: sl, torrentMenuSelected: menu }))
                        }}/>
                    </div>
                    { this.state.selectedItems.length >0 ? <div className="torrent-info" style={{minHeight: this.state.infoSize + "px",maxHeight: this.state.infoSize + "px"}}>
                        <div className="torrent-menu">
                            <div className="buttons-block">
                                <button disabled={this.state.torrentMenuSelected == 0 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(0)}>Info</button>
                                <button disabled={this.state.torrentMenuSelected == 1 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(1)}>Files</button>
                                <button disabled={this.state.torrentMenuSelected == 2 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(2)}>Peers</button>
                            </div>
                            <div onMouseDown={this.extendInfoEvent} className="size-scroller"></div>
                        </div>
                        <div className="torrent-body">
                            {this.state.torrentMenuSelected == 0 ? <InfoTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                            {this.state.torrentMenuSelected == 1 ? <FilesTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                            {this.state.torrentMenuSelected == 2 ? <PeersTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                        </div>
                    </div> : ""}
                    <div className="foot-bar">
                        <div className="speed">
                            <span><img src={Download} alt=""/>{this.state.overallDownloadSpeed}</span>
                            <span><img className="upload" src={Download} alt=""/>{this.state.overallUploadSpeed}</span>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

export default App
