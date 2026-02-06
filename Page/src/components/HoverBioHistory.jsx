import React, {useEffect, useState} from 'react';
import axios from "axios";

function HoverBioHistory(props) {
    const [data,setData] = useState([]);
    useEffect(() => {
        axios.get("/api/history/bio?mid=" + props.uid).then((response) => {
            setData(response.data.data);
        })
    },[])
    return (
        <div>
            {data.map((item, index) => {
                return (
                    <p className={'mt-3'}>{item.Bio}</p>
                )
            })}
        </div>
    );
}

export default HoverBioHistory;