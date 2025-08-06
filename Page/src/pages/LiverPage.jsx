import React from 'react';
import {useNavigate, useParams} from "react-router-dom";
import axios from "axios";
import {LineChart} from "@mui/x-charts/LineChart";
import {Avatar, Card, CardBody, CardHeader, Image, Tooltip} from "@heroui/react";
import HoverMedals from "../components/HoverMedals";

function LiverPage(props) {
    const [fansChart, setFansChart] = React.useState([]);

    const [guardChart, setGuardChart] = React.useState([]);

    const [space, setSpace] = React.useState({});

    const [lives, setLives] = React.useState([]);




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
            var map = [19998, 1998, 138]
            response.data.data.forEach(element => {
                dst.push({
                    UpdatedAt: element.UpdatedAt,
                    Guard: element.Guard.split(",").reduce((a, b) => parseInt(a) + parseInt(b)),
                })
            })
            console.log(dst)
            setGuardChart(dst)
        })
        axios.get(`${protocol}://${host}:${port}/api/live?uid=${id}&limit=1000`).then((response) => {
            setLives(response.data.lives);
        })
    }, [])
    return (
        <div>
            <div className={'flex flex-row mb-6 justify-center '}>
                <Avatar src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${id}`}
                        size={'lg'} onClick={() => {
                    toSpace(id)
                }}/>
                <div className={'ml-4 flex flex-col'}>
                    <span className={'font-bold'}>{space.UName}</span>
                    {space.Verify &&
                        <div className={'flex items-center'}>
                            <svg style={{color: '#ffcc00'}} width="16" height="16" viewBox="0 0 16 16" fill="none"
                                 xmlns="http://www.w3.org/2000/svg">
                                <path
                                    d="M16 8C16 12.4183 12.4183 16 8 16C3.58172 16 0 12.4183 0 8C0 3.58172 3.58172 0 8 0C12.4183 0 16 3.58172 16 8Z"
                                    fill="currentColor"></path>
                                <path fill-rule="evenodd" clip-rule="evenodd"
                                      d="M7.28832 12.7244C7.20127 12.767 7.1148 12.7988 7.02538 12.8C6.80863 12.8042 6.62919 12.6296 6.62564 12.4101C6.62742 12.3717 6.63512 12.3351 6.64814 12.2997L7.40676 8.78586L4.26392 8.79186C4.03651 8.79545 3.85825 8.6209 3.85352 8.40196C3.85588 8.27299 3.9228 8.15362 4.03118 8.08524L8.72206 3.2901C8.80852 3.23732 8.90149 3.20133 8.99743 3.20013C9.21477 3.19653 9.39303 3.37108 9.39776 3.59063C9.39599 3.65541 9.37822 3.71839 9.34506 3.77358L8.59118 7.23047H11.7388C11.9614 7.22687 12.1403 7.40142 12.1444 7.62096C12.1426 7.75113 12.0757 7.8705 11.9668 7.93888L7.28832 12.7244Z"
                                      fill="white"></path>
                            </svg>
                            <span className={'font-bold ml-2'}>{space.Verify}</span>
                        </div>
                    }

                    <span className={'font-thin text-sm'}>{space.Bio}</span>
                </div>
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
                <Tooltip content={<HoverMedals ruid={id}/>}>
                    <div
                        className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝牌<br/>
                        <span
                            className="font-semibold">{space.Medal}</span>
                    </div>
                </Tooltip>
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-2 w-full'}>
                <LineChart
                    xAxis={[{
                        data: fansChart.map((item, index) => (new Date(item.CreatedAt).getTime())),
                        label: "粉丝",
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
                    xAxis={[{
                        data: guardChart.map((item, index) => (new Date(item.UpdatedAt).getTime())),

                        label: "大航海",
                        valueFormatter: (timestamp) => {
                            const date = new Date(timestamp)
                            return date.toLocaleDateString()
                        }
                    }]}
                    series={[
                        {
                            data: guardChart.map((item, index) => (new Date(item.Guard).getTime())),


                        },
                    ]}
                    height={300}
                />
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-6'}>
                {lives.map((live, index) => (
                    <LiveStatisticCard item={live}/>
                ))}
            </div>
        </div>
    );
}


function LiveStatisticCard(props) {
    const redirect = useNavigate()
    var item = props.item
    console.log(item)
    return (
        <div onClick={e => {
            redirect('/lives/' + item.ID)
        }}>
            <Card className={'my-4 mx-2'} isHoverable >
                <CardHeader className="flex-col items-start">
                    <p className="text-large uppercase font-bold">{item.Title}</p>
                    <small className="text-default-500">{new Date(item.CreatedAt).toLocaleString()}</small>
                    <div className={'flex-row flex items-center justify-center'}>
                        <img src={'https://github.com/ant-design/ant-design-icons/raw/refs/heads/master/packages/icons-svg/svg/outlined/message.svg'} style={{width:'20px',height:'20px'}} />
                        <h3 className={'mt-1 ml-0.5'}>{item.Message}</h3>
                        <img src={'https://github.com/ant-design/ant-design-icons/raw/refs/heads/master/packages/icons-svg/svg/outlined/money-collect.svg'} style={{width:'20px',height:'20px'}} className={'ml-1'} />
                        <span className=" mt-1 ml-0.5">{item.Money}</span>
                    </div>
                </CardHeader>
                <CardBody className="overflow-visible py-2">
                    <Image
                        alt="Card background"
                        className="object-cover rounded-xl"
                        src={item.Cover}
                        isBlurred
                        isZoomed

                    />
                </CardBody>
            </Card>
        </div>

    )
}

export default LiverPage;