import React, {FunctionComponent, MouseEvent, useState} from 'react';
import ReactDOM from 'react-dom';
import FilterCheck from "../assets/images/icons/check.svg";

export interface FilterProps {
    onChanged: (filter: string) => void;
}

export const FiltersMenu: FunctionComponent<FilterProps> = (props: FilterProps) => {
    const [selected, setSelected] = useState(0);

    const renderTorrentsList = () => {
        let items: JSX.Element[] = [];
        ["All", "Active","Inactive","Downloading","Seeding","Failed"].forEach((str, i) => {
            const menuClick = () => {
                setSelected(i);
                props.onChanged(str);
            }

            if (selected == i) {
                items.push(<button className={"menu-item selected"} onClick={menuClick}>
                    <img style={{width: "6%"}} src={FilterCheck} alt=""/>
                    <span style={{width:"85%", textAlign:"center", marginRight:"6%"}}>{str}</span>
                </button>)
            } else {
                items.push(<button className={"menu-item not-selected"} onClick={menuClick}>
                    <span style={{width:"100%", textAlign:"center"}}>{str}</span>
                </button>)
            }
        })
        return items;
    }

    return (<div className="filters-menu">
            <h6>Torrents</h6>
            <hr/>
            {renderTorrentsList()}
        </div>);
    };
