import React, {useEffect} from 'react';
import {addToast, Alert, Button, Chip, Input, ToastProvider} from "@heroui/react";
import axios from "axios";

const UserCard = ({
                      mid,
                      name,
                      description,
                  }) => {
    return (
        <div
            className="flex flex-row mt-2 w-full overflow-hidden"
            onClick={() => {
                window.open('https://space.bilibili.com/' + mid)
            }}
        >
            <img
                src={`https://workers.vrp.moe/bilibili/avatar/${mid}`}
                className="w-[80px] h-[80px] rounded-full flex-shrink-0"
                alt={name}
            />

            < div className="flex flex-col ml-2 flex-1 min-w-0 justify-center" >
                <p className="font-bold text-[16px] truncate hover:text-[#00a1d6] transition-colors duration-180">{name}</p>
                <p className="text-[12px] text-[#61666D] truncate">
                    {description}
                </p>
            </div >
        </div >
    );
};

function FansPage() {
    const [uid,setUid] = React.useState('');
    const [remain,setRemain] = React.useState(0);
    const refreshTime = async () => {
        const res = await axios.get("/api/relations/amount")
        setRemain(res.data.amount??0)
    }

    useEffect(() => {
        refreshTime()
    },[])

    const [data,setData] = React.useState([])

    return (
        <div className={'flex flex-col'}>
            <ToastProvider placement={'top-right'}/>
            <div className={'flex flex-row items-center'}>
                <Input className={'max-w-xs'} onValueChange={(e) => {
                    setUid(e.replace('UID:',''))
                }} label={`UID (${remain} Capacity Remain)`} value={uid}></Input>
                <Button className={'ml-3'} onClick={() => {
                    if (remain <= 0) {
                        addToast({
                            title: "Error",
                            description:  "No Capacity  Remain",
                            color: 'danger',
                        })
                    } else {
                        axios.get("/api/relations/fans?uid=" + uid).then((res) => {
                            setData(res.data.data)
                        })
                    }
                    setTimeout(() => {
                        refreshTime()
                    },1000)
                }}>Search</Button>
            </div>
            <div className={'grid grid-cols-1 gap-4 p-4 sm:grid-cols-3'}>
                {data.map((e) => {
                    return (
                        <UserCard mid={e.UID} name={e.UName} description={e.Bio}/>
                    )
                })}
            </div>
        </div>
    );
}

export default FansPage;