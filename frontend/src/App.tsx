import {useState, MouseEvent} from 'react';
import Logo from "./assets/images/logo.svg"
import FilterCheck from "./assets/images/icons/check.svg"
import './tooltip.css';
import './modal.scss';
import {refresh, Table} from "./components/Table";
import {useModal} from "./components/Modal";
import {AddTorrentModal} from "./components/ModalAddTorrent";
import {RemoveTorrent, SetActive} from "../wailsjs/go/main/App";
import {FiltersMenu} from "./components/FiltersMenu";

function App() {
    let selectedItems: string[] = [];
    const [infoSize, infoSetSize] = useState(193);
    const [tableFilter, setTableFilter] = useState({
        type: "all",
        search: "",
    });

    window.addEventListener('resize', () => {
        let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height;
        let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;
        let minINF = document.getElementsByClassName("torrent-info")![0].getBoundingClientRect();

        let minH = minINF.height;
        if (minH > window.innerHeight-(topH+botH)) {
            infoSetSize(window.innerHeight-(topH+botH));
        }
    })

    const extendInfoEvent = (mouseDownEvent: MouseEvent) => {
        const startSize = infoSize;
        const startPosition = { x: mouseDownEvent.pageX, y: mouseDownEvent.pageY };

        function onMouseMove(mouseMoveEvent: any) {
            let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height;
            let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;

            let sz = startSize + startPosition.y - mouseMoveEvent.pageY;
            if (sz > window.innerHeight-(topH+botH)) {
                sz = window.innerHeight-(topH+botH)
            }

            infoSetSize(sz);
        }
        function onMouseUp() {
            document.body.removeEventListener("mousemove", onMouseMove);
            // uncomment the following line if not using `{ once: true }`
            // document.body.removeEventListener("mouseup", onMouseUp);
        }
        document.body.addEventListener("mousemove", onMouseMove);
        document.body.addEventListener("mouseup", onMouseUp, { once: true });
    };

    const addTorrent = useModal();

    return (
        <div id="App">
            {addTorrent.isShown ? <AddTorrentModal onExit={addTorrent.toggle} /> : null}
            <span id="tip" className="tooltiptext">Downloading</span>
            <div className="left-bar">
                <div className="logo-block">
                    <img className="logo-img" src={Logo} alt=""/>
                    <label className="logo-text">TON Torrent</label>
                </div>
                <div className="menu-block">
                    <FiltersMenu onChanged={(v)=>{
                        setTableFilter({
                            type: v,
                            search: tableFilter.search
                        })
                    }}/>
                    <hr/>
                    <div className="actions-menu">
                        <button className="menu-item main" onClick={addTorrent.toggle}>
                            Add Torrent
                        </button>
                        <button className="menu-item">
                            Create Torrent
                        </button>
                        <button className="menu-item">
                            Settings
                        </button>
                    </div>
                    <hr/>
                </div>
                <div className="version-block">
                    <div className="dev-info">
                        Developed with ❤️ by Tonutils team.<br/>
                        Supported by TON Foundation.
                    </div>
                    <div className="ver-info">
                        <span>v1.0.0</span><div className="updates">Check updates</div>
                    </div>
                </div>
            </div>
            <div className="right-screen">
                <div className="top-bar">
                    <div className="top-buttons-container">
                        <button className="top-button start" onClick={()=>{
                            selectedItems.forEach((t) => {
                                SetActive(t, true).then(refresh)
                            })
                        }}>{'>'}</button>
                        <button className="top-button stop" onClick={()=>{
                            selectedItems.forEach((t) => {
                                SetActive(t, false).then(refresh)
                            })
                        }}>D</button>
                        <button className="top-button remove" onClick={()=>{
                            selectedItems.forEach((t) => {
                                RemoveTorrent(t, true).then(refresh)
                            })
                        }}>X</button>
                    </div>
                    <input type="text" className="search-input" placeholder="Search..." onChange={(e)=> {
                        setTableFilter({
                            type: tableFilter.type,
                            search: e.target.value,
                        })
                    }}/>
                </div>
                <div className="torrents-table" style={{height:"50%",maxWidth: '100%', overflowX: "auto"}}>
                    <Table filter={tableFilter} onSelect={(sl)=>{selectedItems=sl}}/>
                </div>
                <div className="torrent-info" style={{minHeight: infoSize+"px"}}>
                    <div className="torrent-menu">

                        <div onMouseDown={extendInfoEvent} className="size-scroller"></div>
                    </div>
                </div>
                <div className="foot-bar">
                    <div className="speed">
                        <span>10 Kb/s</span>
                        <span>210 Mb/s</span>
                    </div>
                </div>
                </div>
        </div>
    )
}

export default App
