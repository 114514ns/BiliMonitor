import React, {useEffect, useState} from 'react';
import axios from "axios";
import {Avatar, Chip} from "@heroui/react";

function HoverMedals(props) {
    var mid = props.mid
    var [data,setData] = useState([])
    useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/medals?mid=${mid}`).then((response) => {
            if (response.data.list == null) {
                response.data.list = []
            }
            response.data.list.sort((a,b) => {
                return a.Score < b.Score
            })
            setData(response.data.list)
        })
    },[])
    return (

        <div className={'flex flex-col overflow-scroll fansMedal overflow-x-hidden'} style={{maxHeight:'600px'}}>
            {data.map((item, index) => (
                <div key={index} className={'mt-2'}>
                    <p className={'font-bold'}>{item.Liver}</p>

                    <div className={'flex flex-row align-middle mt-2'}>
                        <Avatar
                            src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${item.LiverID}`}
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}/>

                        {item.Level ?              <Chip
                            startContent={item.Type?<img  src={getGuardIcon(item.Type)} className={'w-6 h-6'}/>:<div/>}
                            variant="faded"
                            onClick={() => {
                                toSpace(item.LiverID);
                            }}
                            style={{background: getColor(item.Level), color: 'white', marginLeft: '8px'}}
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