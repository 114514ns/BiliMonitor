import React from 'react';
import {useParams} from "react-router-dom";
import axios from "axios";
import {LineChart} from "@mui/x-charts/LineChart";
import {useNavigate} from "react-router";
import {Avatar} from "@heroui/react";

function LiverPage(props) {
    const [fansChart,setFansChart] = React.useState([]);

    const [guardChart,setGuardChart] = React.useState([]);

    const [space,setSpace] = React.useState({});



    let {id} = useParams();

    React.useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/chart/fans?uid=${id}`).then((response) => {
            setFansChart(response.data.data);
        })
        axios.get(`${protocol}://${host}:${port}/api/liver/space?uid=${id}`).then((response) => {
            setSpace(response.data);
        })
        axios.get(`${protocol}://${host}:${port}/api/chart/guard?uid=${id}`).then((response) => {
            var dst = []
            var map = [19998,1998,138]
            response.data.data.forEach(element => {
                dst.push({
                    UpdatedAt:element.UpdatedAt,
                    Guard:element.Guard.split(",").reduce((a, b) => parseInt(a)+parseInt(b)),
                })
            })
            console.log(dst)
            setGuardChart(dst)
        })
    },[])
    return (
        <div>
            <div className={'flex w-full items-center flex-col mb-6'}>
                <span className={'font-bold'}>{space.UName}</span>
                <Avatar src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${id}`} size={'lg'}                onClick={() => {
                    toSpace(id)
                }}/>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-green-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝<br/>
                    <span className="font-semibold">{parseInt(space.Fans).toLocaleString()}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">大航海<br/>
                    <span
                        className="font-semibold">{space.Guard}</span>
                </div>
                <div
                    className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">分区<br/>
                    <span
                        className="font-semibold">{space.Area}</span>
                </div>
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-2 w-full'}>
                <LineChart
                    xAxis={[{
                        data: fansChart.map((item, index) => (new Date(item.CreatedAt).getTime())),
                        label:"粉丝",
                        valueFormatter: (timestamp) => {
                            const date = new Date(timestamp)
                            return date.toLocaleDateString()
                        }
                        }]}
                    series={[
                        {
                            data: fansChart.map((item, index) => (item.Fans))
                        },
                    ]}
                    height={300}
                />
                <LineChart
                    xAxis={[{                         data: guardChart.map((item, index) => (new Date(item.UpdatedAt).getTime())),

                        label:"大航海",
                        valueFormatter: (timestamp) => {
                            const date = new Date(timestamp)
                            return date.toLocaleDateString()
                        } }]}
                    series={[
                        {
                            data: guardChart.map((item, index) => (new Date(item.Guard).getTime())),


                        },
                    ]}
                    height={300}
                />
            </div>
        </div>
    );
}

export default LiverPage;