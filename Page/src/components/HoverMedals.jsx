import React, {useEffect, useState} from 'react';
import axios from "axios";
import {Avatar, Chip} from "@heroui/react";

function HoverMedals(props) {
    var mid = props.mid
    var [data,setData] = useState([])
    useEffect(() => {
        if (props.mid) {
            axios.get(`${protocol}://${host}:${port}/api/medals?mid=${mid}`).then((response) => {
                if (response.data.list == null) {
                    response.data.list = []
                }
                response.data.list.sort((a,b) => {
                    return a.Score < b.Score
                })
                setData(response.data.list)
            })
        } else {
            axios.get(`${protocol}://${host}:${port}/api/fansRank?liver=${props.ruid}&size=100&page=1`).then((response) => {
                if (response.data.list == null) {
                    response.data.list = []
                }
                response.data.list.sort((a,b) => {
                    return a.Score < b.Score
                })
                response.data.list.forEach((element) => {
                    element.Liver = element.UName
                    element.LiverID = element.UID
                })
                setData(response.data.list)
            })
        }

    },[])
    return (

        <div className={'flex flex-col overflow-scroll fansMedal overflow-x-hidden'} style={{maxHeight:'600px'}}>
            {data.map((item, index) => (
                <div key={index} className={'mt-2'}>
                    <p className={'font-bold'}>{item.Liver}</p>

                    <div className={'flex flex-row align-middle mt-2'}>
                        <Avatar
                            src={`${AVATAR_API}${item.LiverID}`}
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}/>

                        {item.Level ?              <Chip
                            startContent={item.Type?<img  src={getGuardIcon(item.Type)} className={'w-6 h-6'}/>:<div/>}
                            variant="faded"
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}
                            style={{background: (new Date().getTime() - new Date(item.UpdatedAt))>168*2*3600*1000?'#5762A799':getColor(item.Level), color: 'white', marginLeft: '8px'}}
                        >
                            {item.MedalName}
                            <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {item.Level}
                                                        </span>
                        </Chip>:<></>}
                    </div>
                </div>
            ))}

        </div>
    )
}

export default HoverMedals;