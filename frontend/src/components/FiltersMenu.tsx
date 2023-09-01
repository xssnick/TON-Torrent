import React, {FunctionComponent, useState} from 'react';

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
                    <span style={{width:"100%", textAlign:"left", marginLeft:"16px"}}>{str}</span>
                </button>)
            } else {
                items.push(<button className={"menu-item not-selected"} onClick={menuClick}>
                    <span style={{width:"100%", textAlign:"left", marginLeft:"16px"}}>{str}</span>
                </button>)
            }
        })
        return items;
    }

    return (<div className="filters-menu">
            {renderTorrentsList()}
        </div>);
    };
